#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <regex.h>
#include <curl/curl.h>
#include "cJSON.h"

struct MemoryStruct {
  char *memory;
  size_t size;
};

static size_t WriteMemoryCallback(void *contents, size_t size, size_t nmemb, void *userp) {
  size_t realsize = size * nmemb;
  struct MemoryStruct *mem = (struct MemoryStruct *)userp;

  char *ptr = realloc(mem->memory, mem->size + realsize + 1);
  if(ptr == NULL) {
    return 0;
  }

  mem->memory = ptr;
  memcpy(&(mem->memory[mem->size]), contents, realsize);
  mem->size += realsize;
  mem->memory[mem->size] = 0;

  return realsize;
}

void respond_error(const char* err) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "error", err);
    char *out = cJSON_Print(root);
    printf("Content-Type: application/json\r\n\r\n%s\n", out);
    free(out);
    cJSON_Delete(root);
    exit(0);
}

void respond_error_cli(const char* err) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "error", err);
    char *out = cJSON_Print(root);
    printf("%s\n", out);
    free(out);
    cJSON_Delete(root);
    exit(1);
}

void perform_query(const char* type, const char* value, int is_cgi) {
    char url[512];
    if (strcmp(type, "domain") == 0) {
        regex_t regex;
        if (regcomp(&regex, "^[a-zA-Z0-9.-]+$", REG_EXTENDED)) {
            if (is_cgi) respond_error("internal regex error");
            else respond_error_cli("internal regex error");
        }
        if (regexec(&regex, value, 0, NULL, 0) == REG_NOMATCH) {
            if (is_cgi) respond_error("invalid domain format");
            else respond_error_cli("invalid domain format");
        }
        regfree(&regex);
        snprintf(url, sizeof(url), "https://rdap.org/domain/%s", value);
    } else if (strcmp(type, "ip") == 0) {
        // very basic validation for ipv4
        regex_t regex;
        if (regcomp(&regex, "^[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+(/[0-9]+)?$", REG_EXTENDED)) {
            if (is_cgi) respond_error("internal regex error");
            else respond_error_cli("internal regex error");
        }
        if (regexec(&regex, value, 0, NULL, 0) == REG_NOMATCH) {
            if (is_cgi) respond_error("invalid IP address format");
            else respond_error_cli("invalid IP address format");
        }
        regfree(&regex);
        snprintf(url, sizeof(url), "https://rdap.org/ip/%s", value);
    } else if (strcmp(type, "autnum") == 0) {
        regex_t regex;
        if (regcomp(&regex, "^(AS|as)?[0-9]+$", REG_EXTENDED)) {
            if (is_cgi) respond_error("internal regex error");
            else respond_error_cli("internal regex error");
        }
        if (regexec(&regex, value, 0, NULL, 0) == REG_NOMATCH) {
            if (is_cgi) respond_error("invalid ASN format");
            else respond_error_cli("invalid ASN format");
        }
        regfree(&regex);
        const char *num_part = value;
        if (strncmp(value, "AS", 2) == 0 || strncmp(value, "as", 2) == 0) {
            num_part += 2;
        }
        snprintf(url, sizeof(url), "https://rdap.org/autnum/%s", num_part);
    } else {
        if (is_cgi) respond_error("unknown query type");
        else respond_error_cli("unknown query type");
    }

    CURL *curl_handle;
    CURLcode res;
    struct MemoryStruct chunk;
    chunk.memory = malloc(1);
    chunk.size = 0;

    curl_global_init(CURL_GLOBAL_ALL);
    curl_handle = curl_easy_init();

    struct curl_slist *headers = NULL;
    headers = curl_slist_append(headers, "Accept: application/rdap+json");

    curl_easy_setopt(curl_handle, CURLOPT_URL, url);
    curl_easy_setopt(curl_handle, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, WriteMemoryCallback);
    curl_easy_setopt(curl_handle, CURLOPT_WRITEDATA, (void *)&chunk);
    curl_easy_setopt(curl_handle, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(curl_handle, CURLOPT_USERAGENT, "libcurl-agent/1.0");

    res = curl_easy_perform(curl_handle);

    if(res != CURLE_OK) {
        char err[256];
        snprintf(err, sizeof(err), "curl_easy_perform() failed: %s", curl_easy_strerror(res));
        if (is_cgi) respond_error(err);
        else respond_error_cli(err);
    } else {
        long response_code;
        curl_easy_getinfo(curl_handle, CURLINFO_RESPONSE_CODE, &response_code);
        if (response_code != 200) {
             cJSON *root = cJSON_Parse(chunk.memory);
             char *out;
             if (root != NULL) {
                 out = cJSON_Print(root);
                 cJSON_Delete(root);
             } else {
                 out = strdup("{\"error\": \"Not found or upstream error\"}");
             }
             if (is_cgi) {
                 printf("Content-Type: application/json\r\n\r\n%s\n", out);
             } else {
                 printf("%s\n", out);
             }
             free(out);
             free(chunk.memory);
             curl_slist_free_all(headers);
             curl_easy_cleanup(curl_handle);
             curl_global_cleanup();
             if (!is_cgi) exit(1);
             return;
        }

        cJSON *root = cJSON_Parse(chunk.memory);
        if (root == NULL) {
            if (is_cgi) respond_error("failed to parse json from upstream");
            else respond_error_cli("failed to parse json from upstream");
        } else {
            char *out = cJSON_Print(root);
            if (is_cgi) {
                printf("Content-Type: application/json\r\n\r\n%s\n", out);
            } else {
                printf("%s\n", out);
            }
            free(out);
            cJSON_Delete(root);
        }
    }

    free(chunk.memory);
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl_handle);
    curl_global_cleanup();
}

void parse_query_string(char *query, char **type, char **value) {
    char *token = strtok(query, "&");
    while (token != NULL) {
        char *eq = strchr(token, '=');
        if (eq != NULL) {
            *eq = '\0';
            char *key = token;
            char *val = eq + 1;
            if (strcmp(key, "domain") == 0 || strcmp(key, "ip") == 0 || strcmp(key, "autnum") == 0) {
                *type = key;
                *value = val;
                return;
            }
        }
        token = strtok(NULL, "&");
    }
}

int main(int argc, char **argv) {
    char *gateway = getenv("GATEWAY_INTERFACE");
    char *qs = getenv("QUERY_STRING");

    if (gateway != NULL || qs != NULL) {
        // CGI mode
        if (qs == NULL) {
            respond_error("missing parameter: expected domain, ip, or autnum");
        }
        char *query = strdup(qs);
        char *type = NULL;
        char *value = NULL;
        parse_query_string(query, &type, &value);
        if (type == NULL || value == NULL) {
            free(query);
            respond_error("missing parameter: expected domain, ip, or autnum");
        }
        perform_query(type, value, 1);
        free(query);
    } else {
        // CLI mode
        if (argc != 3) {
            printf("Usage: radp -domain <domain> | -ip <ip> | -autnum <asn>\n");
            return 1;
        }
        char *type = argv[1];
        char *value = argv[2];
        if (strcmp(type, "-domain") == 0) {
            perform_query("domain", value, 0);
        } else if (strcmp(type, "-ip") == 0) {
            perform_query("ip", value, 0);
        } else if (strcmp(type, "-autnum") == 0) {
            perform_query("autnum", value, 0);
        } else {
            printf("Usage: radp -domain <domain> | -ip <ip> | -autnum <asn>\n");
            return 1;
        }
    }

    return 0;
}
