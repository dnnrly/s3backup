S3_BACKUP=${PWD}/s3backup
TEST_DIR=test/test1

AWS_CMD="aws --endpoint http://localstack:4572"

setup() {
    AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test ${AWS_CMD} s3 rm s3://test-bucket --recursive
    AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test ${AWS_CMD} s3api delete-bucket --bucket test-bucket --region eu-west-1
    AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test ${AWS_CMD} s3api create-bucket --bucket test-bucket --region eu-west-1
}

function s3ls () {
    result="$(AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test ${AWS_CMD} s3 ls test-bucket)"
    echo "${result}"
}

function yqAssert() {
    val="$(yq.v2 read $1 $2)"

    return [ "${val}" = "${3}" ]
}

@test "Can run application" {
    run ${S3_BACKUP}
    [ $status -eq 0 ]
}

@test "Scans test directory and creates index file" {
    run $(cd ${TEST_DIR} && ${S3_BACKUP} create-index)
    [ $status -eq 0 ]
    [ -f ${TEST_DIR}/.s3backup.yaml ]
    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.dir1/file1.key)" = "dir1/file1" ]
    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.dir1/subdir1/file3.key)" = "dir1/subdir1/file3" ]
    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.dir1/subdir2/file2.key)" = "dir1/subdir2/file2" ]
    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.dir2/file5.key)" = "dir2/file5" ]
    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.dir2/subdir1/file4.key)" = "dir2/subdir1/file4" ]
    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.file.key)" = "file" ]

    [ "$(yq.v2 read ${TEST_DIR}/.s3backup.yaml files.dir1/file1.hash)" != "null" ]

    [ "$(s3ls | wc -l)" = "7" ]
}
