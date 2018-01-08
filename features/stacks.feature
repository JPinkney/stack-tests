@addon-che @addon
Feature: Che add-on
  Che addon starts Eclipse Che

  Background: Given Minishift-addons repository is cloned

  Scenario: User enables the che add-on
    When executing "minishift addons enable che" succeeds
    Then stdout should contain "Add-on 'che' enabled"
  
  Scenario: User starts Minishift
    Given Minishift has state "Does Not Exist"
    When executing "minishift start --memory 4GB" succeeds
    Then Minishift should have state "Running"
    And stdout should contain "Che installed"
  
  Scenario Outline: User starts workspace, imports projects, checks run commands
    Given Minishift has state "Running"
    When starting a workspace with stack "<stack_name>" succeeds
    Then workspace start should be successful
    When user runs commands
    Then command should be ran successfully
    
    Examples:
    | stack_name |
    | test       |
  
  Scenario: User stops and deletes the Minishift instance
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then Minishift should have state "Stopped"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"