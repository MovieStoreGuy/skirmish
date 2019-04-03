# Skirmish
_A game day orchestration tool_  
[![Build Status](https://travis-ci.org/MovieStoreGuy/skirmish.svg?branch=master)](https://travis-ci.org/MovieStoreGuy/skirmish)
[![Maintainability](https://api.codeclimate.com/v1/badges/5d181fc764294819aa0b/maintainability)](https://codeclimate.com/github/MovieStoreGuy/skirmish/maintainability)
[![Go Report Card](https://goreportcard.com/badge/github.com/MovieStoreGuy/skirmish)](https://goreportcard.com/report/github.com/MovieStoreGuy/skirmish)
[![Docker Repository on Quay](https://quay.io/repository/lamoviestoreguy/skirmish/status "Docker Repository on Quay")](https://quay.io/repository/lamoviestoreguy/skirmish)  
This application allows for users to run scripted game day events and be able to restore services if required.

## Usage
In order to run skirmish, you need to follow the steps [here](https://developers.google.com/accounts/docs/application-default-credentials) so that the application can assume a role.
If the skirmish is intended to run across multiple projects, then the account will need to have the correct permissions in each one. 

To start using skirmish, it is as simple as:
```sh 
skirmish --plan-path path/to/plan.yml
```
Considering a skirmish can run for over several hours, it is not recommend running within a CI environment that has timed usage.  

**Note: _Skirmish has checks inbuilt to ensure it can restore services if repairable but it makes no guarantees if it receives a SIGKILL._**

An example of a game day plan would be:
```yaml
mode: repairable
projects: # projects defined here ensure the steps will fail if they are mistyped or should be part of the game day
    - staging
    - canary
steps:
    - name: Fail random instances
      description: |-
        Ensure that the systems are inplace for ensuring instance count or 
        that the correct procedure and alerts happen.
      operations: 
        - instance
      projects:
        - staging
      exclude:
        wildcards:        # wildcards support regular expressions
          - "data-node*"
          - "demo-server"
        regions:          # regions / zones support prefix matching
          - "us-west"
      wait: "10m"         # wait 10 minutes to restore instances
      sample: 80.0        # each valid instance will have an 80% chance of being paused
    - name: Stop communication of integration platform components
      description: |-
        Ensure that our platform is still operational when the integration pipeline is cut off
        from communicating
      operations: 
        - egress
      projects:
        - canary
      settings:
        network:
          name: data-ingestion
          deny:
            - protocol: "tcp"
              ports:
                - "8080"
                - "443"
                - "80"
      wait: "30m"
    - name: enstil fear in the cold hearted
      description: |- 
        Let the orchestration platform destroy as much of the project as possible to highlight worst case scenario.
        This will automatically recover once the step has reached its wait time.
      operations:
        - instance
        - egress
        - ingress
      projects:
        - staging
        - canary
      settings:
        network:
          # apply to the default work by leaving name blank
          deny:
            - protocol: "tcp"
              ports:
                - "443"
                - "80"
       wait: "20m"
```