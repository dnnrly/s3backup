@test "Can run application" {
    run ./s3backup
    [ $status -eq 0 ]
}