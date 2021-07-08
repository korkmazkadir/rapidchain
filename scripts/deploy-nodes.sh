#!/bin/bash


number_of_nodes=$1
export REGISTRY_ADDRESS=$2
nic=$3

export NODE_HOSTNAME=$(hostname -i)

mkdir -p output

function throttle()
{

   process_index=$1
   pid=$2

   printf -v gminor "%04x" "$process_index"
   

   group_name_suffix="node_${process_index}"

   # Create a net_cls cgroup
   group_name="net_cls:${group_name_suffix}"
   sudo cgcreate -g "${group_name}"

   # Set the class id for the cgroup
   # By default gmajor is 1
   echo_cmd="echo 0x1${gminor} > /sys/fs/cgroup/net_cls/${group_name_suffix}/net_cls.classid"
   sudo sh -c  "${echo_cmd}"

   # Classify packets from pid into cgroup
   sudo cgclassify -g "${group_name}" "${pid}"

    # adds tasks to specific cgroup one by one
    for task_folder in /proc/"${pid}"/task/*; do
        task_id="${task_folder##*/}"
        sudo cgclassify -g "${group_name}" "${task_id}"
    done

    # By default gmajor is 1
    printf -v class_id "1:%x" "$process_index"

    sudo tc class add dev $nic parent 1: classid "${class_id}" htb rate 20mbit

    sudo tc qdisc add dev $nic parent "${class_id}" netem delay 50ms

}

#Delete previous control groups
sudo cgdelete -r net_cls:/

#Defines network interface to apply tc rules
#nic="eno1"

#Delete previous tc rules
sudo tc qdisc del dev $nic root


#Adds root qdisc
sudo tc qdisc add dev $nic root handle 1: htb
sudo tc filter add dev $nic parent 1: handle 1: cgroup


# tc -s -d class show dev lo

for (( i=1; i<=$number_of_nodes; i++ ))
do

   node_id=${HOSTNAME}_${i}
   nohup ./node 2> output/"$i.log" &
   node_pid=$!

   throttle $i $node_pid

   echo $node_pid >> process.pids
done