*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***

${robot_accounts_xpath}  xpath=//project-detail/nav/ul/li[7]/a
${new_robot_account_xpath}  xpath=//app-robot-account/div/div[2]/clr-dg-action-bar/button
${input_token_name_xpath}  //*[@id="robot_name"]
${input_token_desc_xpath}  //*[@id="robot_desc"]
${save_robot_account_xpath}  xpath=//app-robot-account/div/add-robot/clr-modal[1]/div/div[1]/div/div/div[3]/button[2]
${token_in_creation_success_xpath}  //app-robot-account/div/add-robot/clr-modal[2]/div/div[1]/div/div/div[2]/section/div[2]/hbr-copy-input/div/div[2]/span[1]/input
${copy_robot_account_xpath}  xpath=//app-robot-account/div/add-robot/clr-modal[2]/div/div[1]/div/div/div[2]/section/div[2]/hbr-copy-input/div/div[2]/span[3]/clr-icon
${robot_account_action_xpath}  xpath=//app-robot-account/div/div[2]/clr-dg-action-bar/clr-dropdown/span
${robot_account_delete_xpath}  xpath=//app-robot-account/div/div[2]/clr-dg-action-bar/clr-dropdown/clr-dropdown-menu/button[1]
