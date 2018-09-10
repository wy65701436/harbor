*** Keywords ***
Setup API Test
    ${rc}  ${output}=  Run And Return Rc And Output  make swagger_client 
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
Harbor API Test 
    [Arguments]  ${testcase_name}
    ${current_dir}=  Run  pwd
    Log To Console  ${current_dir}
    ${rc}  ${output}=  Run And Return Rc And Output  SWAGGER_CLIENT_PATH=/home/travis/gopath/src/github.com/goharbor/harbor/harborclient HARBOR_HOST=${ip} python ${testcase_name}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0