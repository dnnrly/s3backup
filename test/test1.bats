S3_BACKUP="${PWD}/s3backup --config ${PWD}/test/config.yaml"
TEST_DIR=${PWD}/test/test1

export PATH=${PWD}/scripts:${GOPATH}/bin:${PATH}

setup() {
    rm -rf ${TEST_DIR}/.s3backup.yaml ${TEST_DIR}/dir1 ${TEST_DIR}/dir1 ${TEST_DIR}/file
    localstack-s3 s3 rm s3://${bucket} --recursive || true

    createTestFiles
}

createTestFiles() {
    mkdir -p ${TEST_DIR}/dir1
    mkdir -p ${TEST_DIR}/dir1/subdir1
    mkdir -p ${TEST_DIR}/dir1/subdir2
    mkdir -p ${TEST_DIR}/dir2
    mkdir -p ${TEST_DIR}/dir2/subdir1

    touch ${TEST_DIR}/dir1/file1
    touch ${TEST_DIR}/dir1/subdir1/file3
    touch ${TEST_DIR}/dir1/subdir2/file2
    touch ${TEST_DIR}/dir2/file5
    touch ${TEST_DIR}/dir2/subdir1/file4
    touch ${TEST_DIR}/file
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

@test "Updloads all files to S3" {
    run $(cd ${TEST_DIR} && ${S3_BACKUP})

    [ "$(localstack-s3 s3 ls test-bucket --recursive | wc -l)" = "7" ]
}

@test "Updloads updated files to S3 without duplicates" {
    cd ${TEST_DIR}
    ${S3_BACKUP} || true
    
    echo "some text" > ${TEST_DIR}/dir2/file-extra
    echo "more text" >> ${TEST_DIR}/dir1/file1

    run $(cd ${TEST_DIR} && ${S3_BACKUP})

    [ "$(echo ${output} | grep -c Uploading)" = "2" ]
    [ "$(localstack-s3 s3 ls test-bucket --recursive | wc -l)" = "8" ]
}
