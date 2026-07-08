<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RDAP & ASN Map Query (PHP)</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .description { background-color: #f0f0f0; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>RDAP & ASN Map Query</h1>

    <div class="description">
        <h2>How to use this tool:</h2>
        <p>This tool allows you to perform RDAP (Registration Data Access Protocol) and ASN (Autonomous System Number) lookups. This PHP version queries the data using backend Go binaries.</p>
        <ul>
            <li><strong>Domain (RDAP):</strong> Enter a domain name (e.g., <code>example.com</code>) to get registration data.</li>
            <li><strong>IP (RDAP):</strong> Enter an IP address (e.g., <code>8.8.8.8</code>) to find out who owns it.</li>
            <li><strong>Autnum (ASN) (RDAP):</strong> Enter an ASN (e.g., <code>AS15169</code>) to get registration details for the autonomous system.</li>
            <li><strong>ASN to IP Map (RADB):</strong> Enter an ASN (e.g., <code>AS15169</code>) to retrieve the list of IP prefixes associated with it from RADB.</li>
        </ul>
    </div>

    <form id="rdapForm">
        <label for="type">Query Type:</label>
        <select id="type" name="type">
            <option value="domain">Domain (RDAP)</option>
            <option value="ip">IP (RDAP)</option>
            <option value="autnum">Autnum (ASN) (RDAP)</option>
            <option value="asnmap">ASN to IP Map (RADB)</option>
        </select>
        <br><br>
        <label for="value">Value:</label>
        <input type="text" id="value" name="value" required placeholder="example.com, 8.8.8.8, AS15169">
        <br><br>
        <button type="submit">Query</button>
    </form>
    <hr>
    <h2>Result:</h2>
    <pre id="result"></pre>

    <script>
        document.getElementById('rdapForm').addEventListener('submit', function(e) {
            e.preventDefault();
            const type = document.getElementById('type').value;
            const value = document.getElementById('value').value;

            const url = `api.php?type=${encodeURIComponent(type)}&value=${encodeURIComponent(value)}`;

            const resultElement = document.getElementById('result');
            resultElement.textContent = 'Loading...';

            fetch(url)
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! status: ${response.status}`);
                    }
                    return response.json();
                })
                .then(data => {
                    resultElement.textContent = JSON.stringify(data, null, 2);
                })
                .catch(error => {
                    resultElement.textContent = 'Error: ' + error.message;
                });
        });
    </script>
</body>
</html>
