Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Keywords ***

Goto Project Robot Accounts
    Retry Element Click  ${robot_accounts_xpath}
    Wait Until Element Is Visible And Enabled  ${new_robot_account_xpath}

Create A Robot Account
    [Arguments]  ${token_name}  ${token_desc}
    Retry Element Click  ${new_robot_account_xpath}
    Wait Until Element Is Visible And Enabled  ${input_token_name_xpath}
    #input necessary info
    Input Text  xpath=${input_token_name_xpath}  ${token_name}
    Input Text  xpath=${input_token_desc_xpath}  ${token_desc}
    Retry Element Click  ${save_robot_account_xpath}
    Wait Until Element Is Visible And Enabled  ${token_in_creation_success_xpath}
    ${token}=  Get Text  xpath=${token_in_creation_success_xpath}
    Log To Console  ${token}
    Retry Element Click  ${copy_robot_account_xpath}
    [Return]  ${token}

Disable A Robot Account
    [Arguments]  ${token_name}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${robotname}')]//clr-checkbox-wrapper//label
    Retry Element Click  ${robot_account_action_xpath}
    Retry Element Click  ${robot_account_delete_xpath}
