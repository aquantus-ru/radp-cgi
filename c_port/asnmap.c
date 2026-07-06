#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <regex.h>
#include <ctype.h>
#include "cJSON.h"

#define RADB_HOST "whois.radb.net"
#define RADB_PORT 43

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

int connect_to_radb(const char **err_msg) {
    int sockfd;
    struct sockaddr_in serv_addr;
    struct hostent *server;

    sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) {
        *err_msg = "failed to open socket";
        return -1;
    }

    server = gethostbyname(RADB_HOST);
    if (server == NULL) {
        *err_msg = "failed to resolve whois.radb.net";
        close(sockfd);
        return -1;
    }

    memset((char *) &serv_addr, 0, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;
    memcpy((char *)&serv_addr.sin_addr.s_addr, (char *)server->h_addr, server->h_length);
    serv_addr.sin_port = htons(RADB_PORT);

    // Basic timeout implementation
    struct timeval tv;
    tv.tv_sec = 10;
    tv.tv_usec = 0;
    setsockopt(sockfd, SOL_SOCKET, SO_RCVTIMEO, (const char*)&tv, sizeof tv);
    setsockopt(sockfd, SOL_SOCKET, SO_SNDTIMEO, (const char*)&tv, sizeof tv);

    if (connect(sockfd, (struct sockaddr *) &serv_addr, sizeof(serv_addr)) < 0) {
        *err_msg = "failed to connect to whois.radb.net";
        close(sockfd);
        return -1;
    }

    return sockfd;
}

// Very simplistic read string until \n
char* read_line(int fd) {
    char *buf = malloc(1024);
    int capacity = 1024;
    int pos = 0;
    char c;
    while(read(fd, &c, 1) == 1) {
        if (c == '\n') break;
        buf[pos++] = c;
        if (pos == capacity) {
            capacity *= 2;
            buf = realloc(buf, capacity);
        }
    }
    buf[pos] = '\0';
    // Trim \r if present
    if (pos > 0 && buf[pos-1] == '\r') {
        buf[pos-1] = '\0';
    }
    return buf;
}

cJSON* fetch_radb(int sockfd, const char *query, const char **err_msg) {
    char send_buf[256];
    snprintf(send_buf, sizeof(send_buf), "%s\n", query);

    if (write(sockfd, send_buf, strlen(send_buf)) < 0) {
        *err_msg = "failed to write to socket";
        return NULL;
    }

    char *line = read_line(sockfd);
    if (strlen(line) == 0) {
        free(line);
        *err_msg = "empty response from radb";
        return NULL;
    }

    if (strcmp(line, "D") == 0) {
        free(line);
        return cJSON_CreateArray(); // Empty prefixes
    }

    if (line[0] == 'A') {
        free(line);
        char *data_line = read_line(sockfd);
        char *c_line = read_line(sockfd); // read 'C'
        free(c_line);

        cJSON *arr = cJSON_CreateArray();
        char *token = strtok(data_line, " \t");
        while (token != NULL) {
            cJSON_AddItemToArray(arr, cJSON_CreateString(token));
            token = strtok(NULL, " \t");
        }
        free(data_line);
        return arr;
    }

    if (line[0] == 'F') {
        static char err_buf[256];
        snprintf(err_buf, sizeof(err_buf), "radb error: %s", line);
        *err_msg = err_buf;
        free(line);
        return NULL;
    }

    static char unexp_buf[256];
    snprintf(unexp_buf, sizeof(unexp_buf), "unexpected response from radb: %s", line);
    *err_msg = unexp_buf;
    free(line);
    return NULL;
}

void perform_asnmap(const char *asn_in, int is_cgi) {
    char asn[254];
    strncpy(asn, asn_in, sizeof(asn)-1);
    asn[sizeof(asn)-1] = '\0';

    // Normalize: uppercase and trim
    for (int i = 0; asn[i]; i++) {
        asn[i] = toupper(asn[i]);
    }

    // validate
    regex_t regex;
    if (regcomp(&regex, "^(AS)?[0-9]+$", REG_EXTENDED)) {
        if (is_cgi) respond_error("internal regex error");
        else respond_error_cli("internal regex error");
    }
    if (regexec(&regex, asn, 0, NULL, 0) == REG_NOMATCH) {
        if (is_cgi) respond_error("invalid ASN format: expected AS<number> or <number>");
        else respond_error_cli("invalid ASN format: expected AS<number> or <number>");
    }
    regfree(&regex);

    char full_asn[256];
    char num_part[256];
    if (strncmp(asn, "AS", 2) == 0) {
        strcpy(full_asn, asn);
        strcpy(num_part, asn + 2);
    } else {
        snprintf(full_asn, sizeof(full_asn), "AS%s", asn);
        strcpy(num_part, asn);
    }

    const char *err_msg = NULL;
    int fd_v4 = connect_to_radb(&err_msg);
    if (fd_v4 < 0) {
        if (is_cgi) respond_error(err_msg);
        else respond_error_cli(err_msg);
    }

    char query_v4[256];
    snprintf(query_v4, sizeof(query_v4), "!gAS%s", num_part);
    cJSON *v4_prefixes = fetch_radb(fd_v4, query_v4, &err_msg);
    close(fd_v4);

    if (v4_prefixes == NULL) {
        if (is_cgi) respond_error(err_msg);
        else respond_error_cli(err_msg);
    }

    int fd_v6 = connect_to_radb(&err_msg);
    cJSON *v6_prefixes = NULL;
    if (fd_v6 >= 0) {
        char query_v6[256];
        snprintf(query_v6, sizeof(query_v6), "!6AS%s", num_part);
        v6_prefixes = fetch_radb(fd_v6, query_v6, &err_msg);
        close(fd_v6);
    }

    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "asn", full_asn);

    // Merge arrays into one
    if (v6_prefixes != NULL) {
        int v6_count = cJSON_GetArraySize(v6_prefixes);
        for (int i = 0; i < v6_count; i++) {
            cJSON *item = cJSON_GetArrayItem(v6_prefixes, i);
            cJSON_AddItemToArray(v4_prefixes, cJSON_CreateString(item->valuestring));
        }
        cJSON_Delete(v6_prefixes);
    }

    cJSON_AddItemToObject(root, "prefixes", v4_prefixes);

    char *out = cJSON_Print(root);
    if (is_cgi) {
        printf("Content-Type: application/json\r\n\r\n%s\n", out);
    } else {
        printf("%s\n", out);
    }

    free(out);
    cJSON_Delete(root);
}

void parse_query_string(char *query, char **value) {
    char *token = strtok(query, "&");
    while (token != NULL) {
        char *eq = strchr(token, '=');
        if (eq != NULL) {
            *eq = '\0';
            char *key = token;
            char *val = eq + 1;
            if (strcmp(key, "asn") == 0) {
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
            respond_error("missing parameter: expected asn");
        }
        char *query = strdup(qs);
        char *value = NULL;
        parse_query_string(query, &value);
        if (value == NULL) {
            free(query);
            respond_error("missing parameter: expected asn");
        }
        perform_asnmap(value, 1);
        free(query);
    } else {
        // CLI mode
        char *asn = NULL;
        for (int i = 1; i < argc; i++) {
            if (strcmp(argv[i], "-asn") == 0 && i + 1 < argc) {
                asn = argv[i+1];
                i++;
            } else if (argv[i][0] != '-') {
                asn = argv[i];
            }
        }

        if (asn == NULL) {
            printf("Usage: asnmap <asn> | -asn <asn>\n");
            return 1;
        }

        perform_asnmap(asn, 0);
    }

    return 0;
}
