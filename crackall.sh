#!/bin/bash

while read line
do
    echo $line
    plaintext=`echo $line | ./playfaircrack crack`
    echo $plaintext
    echo $plaintext >> ./plaintexts.txt

    # echo $line | ./playfaircrack crack -v
done < ./ciphertexts.txt