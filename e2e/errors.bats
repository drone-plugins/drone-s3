#!/usr/bin/env bats

load utils

@test 'Error code propagates' {
    JUST_FAIL=1 \
    run drone_s3 '{}'
    [ "$status" -ne "0" ]
}

@test 'Skipped source relies on awscli error code' {
    JUST_FAIL=1 \
    run drone_s3 '{"vargs":{"source":null}}'
    echo $output
    [ "$status" -ne "0" ]
}

@test 'Skipped target relies on awscli error code' {
    JUST_FAIL=1 \
    run drone_s3 '{"vargs":{"target":null}}'
    echo $output
    [ "$status" -ne "0" ]
}

