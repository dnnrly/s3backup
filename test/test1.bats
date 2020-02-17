S3_BACKUP="${PWD}/s3backup --config ${PWD}/test/config.yaml"
TEST_DIR=test/test1

export PATH=${PWD}/scripts:${GOPATH}/bin:${PATH}

setup() {
    rm -f ${TEST_DIR}/.s3backup.yaml
    localstack-s3 s3 rm s3://test-bucket --recursive
    localstack-s3 s3api delete-bucket --bucket test-bucket --region eu-west-1
    localstack-s3 s3api create-bucket --bucket test-bucket --region eu-west-1
}

@test "Scans test directory and creates index file" {
    run $(cd ${TEST_DIR} && ${S3_BACKUP} create-index)

    [ -f ${TEST_DIR}/.s3backup.yaml ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.dir1/file1.key)" = "dir1/file1" ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.dir1/subdir1/file3.key)" = "dir1/subdir1/file3" ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.dir1/subdir2/file2.key)" = "dir1/subdir2/file2" ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.dir2/file5.key)" = "dir2/file5" ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.dir2/subdir1/file4.key)" = "dir2/subdir1/file4" ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.file.key)" = "file" ]
    [ "$(yq read ${TEST_DIR}/.s3backup.yaml files.dir1/file1.hash)" != "null" ]
}

@test "Updloads to S3" {
    run $(cd ${TEST_DIR} && ${S3_BACKUP})

    [ "$(localstack-s3 s3 ls test-bucket --recursive | wc -l)" = "7" ]
}
