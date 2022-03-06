<?php

if ($_SERVER["REQUEST_METHOD"] != "POST") {
    header("HTTP/1.1 405 Method Not Allowed");
    exit();
}

$body = json_decode(file_get_contents('php://input'), true);

// Do things here

header("HTTP/1.1 201 Created");
exit();
