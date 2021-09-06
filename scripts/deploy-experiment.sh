#!/bin/bash

function deploy {

    experimentConfig="$1"
    
    # removes existing config file
    rm ./artifacts/config.json

    # moves new config file
    mv "$experimentConfig" ./artifacts/config.json

    echo "uploading config..."
    # upload config
    ansible-playbook -i hosts playbooks/upload-config.yml

    echo "deploying experiment..."
    # deploy experiment
    ansible-playbook -i hosts playbooks/deploy-experiment.yml

    echo "waiting for the experiment..."
    # wait for the end of experiment
    ansible-playbook -i hosts playbooks/wait-endof-experiment.yml
}


# install dependencies
# ansible-playbook -i hosts playbooks/install-dependencies.yml

# upload artifacts
# ansible-playbook -i hosts playbooks/upload-artifacts.yml

# iterates over config files
for filename in ./experiments_to_conduct/*.json; do

    echo " deploying $experimentConfig"

    deploy "$filename"

done


# download stats
ansible-playbook -i hosts playbooks/download-stats.yml 




