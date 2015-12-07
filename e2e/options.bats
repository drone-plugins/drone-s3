#!/usr/bin/env bats

load utils

@test "Can specify acl" {
    EXPECTED_CMD='"--acl" "my-acl"' \
    run drone_s3 '{"vargs":{"acl":"my-acl"}}'
    [ "$status" -eq 0 ]
}

@test "Can specify region" {
    EXPECTED_CMD='"--region" "far-far-away"' \
    run drone_s3 '{"vargs":{"region":"far-far-away"}}'
    [ "$status" -eq 0 ]
}

@test "Can specify bucket" {
    EXPECTED_CMD='"s3" "cp" "\S+" "s3://spilt-bucket/' \
    run drone_s3 '{"vargs":{"bucket":"spilt-bucket"}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'Can order recursive' {
    EXPECTED_CMD='"--recursive"' \
    run drone_s3 '{"vargs":{"recursive":true}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'Target with leading slash' {
    EXPECTED_CMD='"s3://[^/]*/a/b/c/d"' \
    run drone_s3 '{"vargs":{"target":"/a/b/c/d"}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'Target without leading slash' {
    EXPECTED_CMD='"s3://[^/]*/a/b/c/d"' \
    run drone_s3 '{"vargs":{"target":"a/b/c/d"}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'Can specify access key' {
    EXPECTED_AWS_ACCESS_KEY_ID='myveryownaccesskey' \
    run drone_s3 '{"vargs":{"access_key":"myveryownaccesskey"}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'Can specify secret key' {
    EXPECTED_AWS_SECRET_ACCESS_KEY='sosecret' \
    run drone_s3 '{"vargs":{"secret_key":"sosecret"}}'
    echo $output
    [ "$status" -eq 0 ]
}

