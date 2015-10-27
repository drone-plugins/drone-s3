#!bash

file_must_be_removed() {
    if [ -f "$1" ]
    then
        echo "aws CLI was not executed"
        return 1
    fi
}

drone_s3() {
    local tmp=$(mktemp ${BATS_TMPDIR}/XXXXXXX)
    echo "$1" | ${BATS_TEST_DIRNAME}/bin/patch_json.go ${BATS_TEST_DIRNAME}/template.json | \
    RM_FILE=${tmp} \
    PATH=${BATS_TEST_DIRNAME}/bin:${PATH} \
    EXPECTED_CMD=${EXPECTED_CMD} UNEXPECTED_CMD=${UNEXPECTED_CMD} \
    EXPECTED_AWS_ACCESS_KEY_ID=${EXPECTED_AWS_ACCESS_KEY_ID} UNEXPECTED_AWS_ACCESS_KEY_ID=${UNEXPECTED_AWS_ACCESS_KEY_ID} \
    EXPECTED_AWS_SECRET_ACCESS_KEY=${EXPECTED_AWS_SECRET_ACCESS_KEY} UNEXPECTED_AWS_SECRET_ACCESS_KEY=${UNEXPECTED_AWS_SECRET_ACCESS_KEY} \
    ${BATS_TEST_DIRNAME}/../drone-s3 && file_must_be_removed ${tmp}
}
