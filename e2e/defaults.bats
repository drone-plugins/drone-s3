#!/usr/bin/env bats

load utils

@test "Default acl is private" {
    EXPECTED_CMD='"--acl" "private"' \
    run drone_s3 '{"vargs":{"acl":null}}'
    [ "$status" -eq 0 ]
}

@test 'Default region is us-east-1' {
    EXPECTED_CMD='"--region" "us-east-1"' \
    run drone_s3 '{"vargs":{"region":null}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'Default is not recursive' {
    UNEXPECTED_CMD='"--recursive"' \
    run drone_s3 '{"vargs":{"recursive":null}}'
    [ "$status" -eq 0 ]
}

@test 'No access key by default' {
    UNEXPECTED_AWS_ACCESS_KEY_ID=x \
    run drone_s3 '{"vargs":{"access_key":null}}'
    echo $output
    [ "$status" -eq 0 ]
}

@test 'No secret key by default' {
    UNEXPECTED_AWS_SECRET_ACCESS_KEY=x \
    run drone_s3 '{"vargs":{"secret_key":null}}'
    echo $output
    [ "$status" -eq 0 ]
}
