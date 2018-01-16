@addon-che @addon
Feature: Che add-on
  Che addon starts Eclipse Che

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
    When starting a workspace with stack "<stack>" succeeds
    Then workspace should have state "Running"
    When importing the sample project "<sample>" succeeds
    Then workspace should have 1 project
    When user runs build command on sample "<sample>"
    Then exit code should be 0
    When user stops workspace
    Then workspace stop should be successful
    When workspace is removed
    Then workspace removal should be successful
    
    Examples:
    | stack                 | sample                                                                   |
    | .NET CentOS           | https://github.com/che-samples/dotnet-web-simple                         |
    | CentOS nodejs         | https://github.com/che-samples/web-nodejs-sample                         |
    | CentOS Wildfly Swarm  | https://github.com/wildfly-swarm-openshiftio-boosters/wfswarm-rest-http  |
    | Eclipse Vert.x        | https://github.com/openshiftio-vertx-boosters/vertx-http-booster         |
    | Java CentOS           | https://github.com/che-samples/console-java-simple                       |
    | Spring Boot           | https://github.com/snowdrop/spring-boot-http-booster                     |
  
  Scenario: User stops and deletes the Minishift instance
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then Minishift should have state "Stopped"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"