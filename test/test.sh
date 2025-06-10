#!/usr/bin/env bash

ENDPOTINT="http://localhost:8081"

# make sure utils command exists
if ! command -v curl &>/dev/null; then
  echo "curl command not found. Please install curl to run this script."
  exit 1
fi

# make sure jq command exists
if ! command -v jq &>/dev/null; then
  echo "jq command not found. Please install jq to run this script."
  exit 1
fi

function GET() {
  curl -s -X GET "$ENDPOTINT/$1"
}

function POST() {
  curl -s -X POST "$ENDPOTINT/$1" -d "$2"
}

# display message in GREEN color function OK() { echo -e "\033[0;32m$1\033[0m" }
function OK() {
  echo -e "\033[0;32m$1\033[0m"
}

# display message in RED color
function FAIL() {
  echo -e "\033[0;31m$1\033[0m"
}

# display message in block color
function BLOCK() {
  echo -e "\033[0;34m$1\033[0m"
}

#### testing function here
function test_ping() {
  BLOCK "Testing ping..."

  body=$(GET "ping")
  # check if response body message field is "ping"
  if [[ $(echo "$body" | jq -r '.message') == "pong" ]]; then
    OK "Ping test passed."
  else
    FAIL "Ping test failed. Response: $body"
    exit 1
  fi
}

function test_templates_list() {
  BLOCK "Testing templates list..."
  body=$(GET "api/templates")

  # first object in the response field name should be "sign"
  if [[ $(echo "$body" | jq -r '.data.[0].name') == "sign" ]]; then
    OK "Templates list test passed."
  else
    FAIL "Templates list test failed. Response: $body"
    exit 1
  fi
}

function test_tempalte_get() {
  BLOCK "Testing template get..."

  body=$(GET "api/templates/1")
  if [[ $(echo "$body" | jq -r '.data.name') == "sign" ]]; then
    OK "Template get test passed."
  else
    FAIL "Template get test failed. Response: $body"
    exit 1
  fi
}

function test_template_default() {
  BLOCK "Testing template default..."

  body=$(GET "api/templates/default")
  # check default template name is "sign"
  if [[ $(echo "$body" | jq -r '.data.name') == "sign" ]]; then
    OK "Template default test passed."
  else
    FAIL "Template default test failed. Response: $body"
    exit 1
  fi
}

function test_movie_create() {
  BLOCK "testing movie create..."
  create_body='{"idea": "Test Movie", "tpl_name": "sign"}'

  body=$(POST "api/movies" "$create_body")
  echo $body
}

function test_movie_list() {
  BLOCK "Testing movie list..."

  body=$(GET "api/movies")
  echo $body
  # body data field should be an array more then 1 items
  if [[ $(echo "$body" | jq -r '.data | length') -gt 0 ]]; then
    OK "Movie list test passed."
  else
    FAIL "Movie list test failed. Response: $body"
    exit 1
  fi
}

function test_movie_get() {
  BLOCK "Testing movie get..."

  body=$(GET "api/movies/1")
  # check if response data field has name "Test Movie"
  if [[ $(echo "$body" | jq -r '.data.idea') == "Test Movie" ]]; then
    OK "Movie get test passed."
  else
    FAIL "Movie get test failed. Response: $body"
    exit 1
  fi
}

function test_movie_generate_script() {
  BLOCK "Testing movie generate script..."

  POST_BODY='{"idea": "列出三个金牛女生的特点， 每一个加以扩展和说明"}'
  body=$(POST "api/movies/1/generate_script" "${POST_BODY}")

  echo $body
}

function test_generate_script_image() {
  BLOCK "Testing movie generate script image..."

  body=$(POST "api/movies/1/generate_image")

  echo $body
}

test_ping
test_templates_list
test_tempalte_get
test_template_default
test_movie_create
test_movie_list
test_movie_get
test_movie_generate_script
test_generate_script_voice
test_generate_script_image
