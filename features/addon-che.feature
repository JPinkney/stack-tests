@che @che5
Feature: Che add-on
  Che addon starts Eclipse Che

  Scenario Outline: User starts workspace, imports projects, checks run commands
    Given Minishift has state "Running"
    When we try to get the stacks information
    Then the stacks should not be empty
    When starting a workspace with stack "<stack>" succeeds
    Then workspace should have state "RUNNING"
    When importing the sample project "<sample>" succeeds
    Then workspace should have 1 project
    When user runs command on sample "<sample>"
    Then exit code should be 0
    When user stops workspace
    Then workspace should have state "STOPPED"
    When workspace is removed
    Then workspace removal should be successful
    
    Examples:
    | stack                 | sample                                                                   |
    | CentOS nodejs         | https://github.com/che-samples/web-nodejs-sample.git                     |
    | Eclipse Vert.x        | https://github.com/openshiftio-vertx-boosters/vertx-http-booster         |
    | Java CentOS           | https://github.com/che-samples/console-java-simple.git                   |
    | Spring Boot           | https://github.com/snowdrop/spring-boot-http-booster                     |