#!/usr/bin/env bash

echo -ne '|                    | (0%) -> noise_text_image\r'
go build -o noise_text_image noise_text_image.go
echo -ne '|#####               | (25%) -> noise_builder_gl\r'
go build -o noise_builder_gl noise_builder_gl.go exampleapp.go
echo -ne '|##########          | (50%) -> noise_from_json_gl\r'
go build -o noise_from_json_gl noise_from_json_gl.go exampleapp.go
echo -ne '|###############     | (75%) -> noise_gl\r'
go build -o noise_gl noise_gl.go exampleapp.go
echo -ne '|####################| (100%) ----------> FINISHED\r'
