# file: $GOPATH/src/godogs/features/godogs.feature
Feature: Stack Tests

  Scenario Outline: Check exec of main binary as default user
    Given we have stack name "<name>" imageName "<imageName>" cmd "<cmd>" expectedOutput "<expectedOutput>" sample "<sample>" and sampleFolderName "<sampleFolderName>"
    When we check exec of main binary as default user
    Then stdout should be "<expectedOutput>"
  Examples:
    |     name          |                 imageName                              | cmd                                                    | expectedOutput        | sample                                                 | sampleFolderName     |
    |  blank-default    |  registry.devshift.net/che/ubuntu_jdk8                 | svn --version                                          | 1.9.3                 | https://github.com/che-samples/console-java-simple.git | console-java-simple  |

  Scenario Outline: Check exec of main binary as arbitrary user
    Given we have stack name "<name>" imageName "<imageName>" cmd "<cmd>" expectedOutput "<expectedOutput>" sample "<sample>" and sampleFolderName "<sampleFolderName>"
    When we check exec of main binary as arbitrary user
    Then stdout should be "<expectedOutput>"
  Examples:
    |     name          |                 imageName                              | cmd                                                    | expectedOutput        | sample                                                 | sampleFoldername     |
    |  blank-default    |  registry.devshift.net/che/ubuntu_jdk8                 | svn --version                                          | 1.9.3                 | https://github.com/che-samples/console-java-simple.git | console-java-simple  |
  
  Scenario Outline: Check run main binary from bash as default user
    Given we have stack name "<name>" imageName "<imageName>" cmd "<cmd>" expectedOutput "<expectedOutput>" sample "<sample>" and sampleFolderName "<sampleFolderName>"
    When we check run main binary from bash as default user
    Then stdout should be "<expectedOutput>"
  Examples:
    |     name          |                 imageName                              | cmd                                                    | expectedOutput        | sample                                                 | sampleFoldername     |
    |  blank-default    |  registry.devshift.net/che/ubuntu_jdk8                 | svn --version                                          | 1.9.3                 | https://github.com/che-samples/console-java-simple.git | console-java-simple  |

  Scenario Outline: Check run main binary from bash as arbitrary user
    Given we have stack name "<name>" imageName "<imageName>" cmd "<cmd>" expectedOutput "<expectedOutput>" sample "<sample>" and sampleFolderName "<sampleFolderName>"
    When we check run main binary from bash as arbitrary user
    Then stdout should be "<expectedOutput>" 
  Examples:
    |     name          |                 imageName                              | cmd                                                    | expectedOutput        | sample                                                 | sampleFoldername     |
    |  blank-default    |  registry.devshift.net/che/ubuntu_jdk8                 | svn --version                                          | 1.9.3                 | https://github.com/che-samples/console-java-simple.git | console-java-simple  |

  Scenario Outline: Check run commands as default user
    Given we have stack name "<name>" imageName "<imageName>" cmd "<cmd>" expectedOutput "<expectedOutput>" sample "<sample>" and sampleFolderName "<sampleFolderName>"
    When we check run commands as default user
    Then stdout should be "<expectedOutput>"
  Examples:
    |     name          |                 imageName                              | cmd                                                    | expectedOutput        | sample                                                 | sampleFoldername     |
    |  blank-default    |  registry.devshift.net/che/ubuntu_jdk8                 | svn --version                                          | 1.9.3                 | https://github.com/che-samples/console-java-simple.git | console-java-simple  |
  
  Scenario Outline: Check run commands as arbitrary user
    Given we have stack name "<name>" imageName "<imageName>" cmd "<cmd>" expectedOutput "<expectedOutput>" sample "<sample>" and sampleFolderName "<sampleFolderName>"
    When we check run commands as arbitrary user
    Then stdout should be "<expectedOutput>"
  Examples:
    |     name          |                 imageName                              | cmd                                                    | expectedOutput        | sample                                                 | sampleFoldername     |
    |  blank-default    |  registry.devshift.net/che/ubuntu_jdk8                 | svn --version                                          | 1.9.3                 | https://github.com/che-samples/console-java-simple.git | console-java-simple  | 