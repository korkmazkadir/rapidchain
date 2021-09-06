#!/bin/bash

macroblock_sizes=(2)
concurrency_constants=(1 2 4 8 16 32)
chunk_count=128

file_index=1
for macroblock_size in "${macroblock_sizes[@]}"
do
    macroblock_size_real=$(($macroblock_size * 1000000))
    for cc in "${concurrency_constants[@]}"
    do
        chunk_count=$((128 / $cc))
        printf -v file_name "%04d_%dMB_CC%d.json" ${file_index} ${macroblock_size} ${cc}
        echo "${file_name}"

        jq --arg bs "$macroblock_size_real" --arg cc "$cc" --arg chunkc "$chunk_count" '.BlockSize =($bs|tonumber) | .LeaderCount =($cc|tonumber) | .BlockChunkCount =($chunkc|tonumber)   ' template_config.json > "./experiments_to_conduct/${file_name}"

        ((file_index++))
    done

done
