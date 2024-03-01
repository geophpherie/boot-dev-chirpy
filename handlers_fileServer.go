package main

import "net/http"

const filepathRoot = "."

var fileServerHandler = http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
