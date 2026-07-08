# RDAP & ASN Map Tool - PHP Version

This PHP application provides a web interface and API to perform RDAP (Registration Data Access Protocol) and ASN lookups. It utilizes the precompiled binaries from the `cgi-app` folder to fetch the data.

## Requirements
- PHP 8.0+
- The compiled Go binaries (`radp` and `asnmap`) must be present in the `cgi-app` directory (one level up from this directory).

## Setup & Installation

1. First, make sure you have compiled the binaries in the `cgi-app/src` directory:
   ```bash
   cd ../cgi-app/src
   ./run_demo.sh
   ```

2. To host this app, you can use any PHP-compatible web server like Apache or Nginx. Simply point the document root to this directory (`php-app`).

3. For a quick test, you can use the built-in PHP development server:
   ```bash
   cd php-app
   php -S localhost:8000
   ```
   Then open `http://localhost:8000` in your web browser.

## Architecture
- `index.php`: Contains the front-end HTML and JavaScript for making queries.
- `api.php`: The backend API that accepts requests, sets up the proper CGI environment variables, calls the underlying Go binaries (`radp` or `asnmap`), and formats the response.
