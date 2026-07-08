<?php
header('Content-Type: application/json');

$type = $_GET['type'] ?? '';
$value = $_GET['value'] ?? '';

if (empty($type) || empty($value)) {
    http_response_code(400);
    echo json_encode(["error" => "Missing type or value parameter"]);
    die();
}

// Path to the precompiled binaries
$radpBin = __DIR__ . '/../cgi-app/radp';
$asnmapBin = __DIR__ . '/../cgi-app/asnmap';

if (!file_exists($radpBin) || !file_exists($asnmapBin)) {
    http_response_code(500);
    echo json_encode(["error" => "Backend binaries not found. Please compile them in cgi-app/src/."]);
    die();
}

$output = '';

// Setup CGI environment variables
$env = [
    'GATEWAY_INTERFACE' => 'CGI/1.1',
    'REQUEST_METHOD' => 'GET',
    'SCRIPT_NAME' => '/cgi-bin/' . ($type === 'asnmap' ? 'asnmap' : 'radp'),
];

if ($type === 'asnmap') {
    $env['QUERY_STRING'] = http_build_query(['asn' => $value]);
    $command = escapeshellcmd($asnmapBin);
} else if (in_array($type, ['domain', 'ip', 'autnum'])) {
    $env['QUERY_STRING'] = http_build_query([$type => $value]);
    $command = escapeshellcmd($radpBin);
} else {
    http_response_code(400);
    echo json_encode(["error" => "Invalid query type"]);
    die();
}

$descriptorspec = [
    0 => ["pipe", "r"],
    1 => ["pipe", "w"],
    2 => ["pipe", "w"]
];

$process = proc_open($command, $descriptorspec, $pipes, null, $env);

if (is_resource($process)) {
    fclose($pipes[0]);

    $output = stream_get_contents($pipes[1]);
    fclose($pipes[1]);

    $errors = stream_get_contents($pipes[2]);
    fclose($pipes[2]);

    $return_value = proc_close($process);
} else {
    http_response_code(500);
    echo json_encode(["error" => "Failed to start binary process"]);
    die();
}

if ($output === null || $output === false || $output === '') {
    http_response_code(500);
    echo json_encode(["error" => "Failed to execute binary or no output", "details" => $errors]);
    die();
}

// The binaries return JSON, but as a CGI script they output "Content-Type: application/json\r\n\r\n" first.
// We need to strip the HTTP headers from the output before echoing it back as JSON.
$parts = explode("\r\n\r\n", $output, 2);
if (count($parts) === 2) {
    echo $parts[1];
} else {
    echo $output; // Fallback just in case
}
