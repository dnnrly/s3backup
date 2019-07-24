S3_BACKUP=${PWD}/s3backup
TETS_DIR=test/test1

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
    [ -f ${TEST_DIR}/s3backup-index.yml ]
    [ "$(yq.v2 ${TEST_DIR}/s3backup-index.yml files.dir1/file1.key)" = "dir1/file1" ]
    [ "$(yq.v2 ${TEST_DIR}/s3backup-index.yml files.dir1/subdir1/file3.key)" = "dir1/subdir1/file3" ]
    [ "$(yq.v2 ${TEST_DIR}/s3backup-index.yml files.dir1/subdir2/file2.key)" = "dir1/subdir2/file2" ]
    [ "$(yq.v2 ${TEST_DIR}/s3backup-index.yml files.dir2/file5.key)" = "dir2/file5" ]
    [ "$(yq.v2 ${TEST_DIR}/s3backup-index.yml files.dir2/subdir1/file4.key)" = "dir1/file1" ]
    [ "$(yq.v2 ${TEST_DIR}/s3backup-index.yml files.file.key)" = "file" ]
}