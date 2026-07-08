#!/usr/bin/env bash

# copy any updated certs from letsencrypt updates

pwd

#for n in {0..5}
for n in {0}
do
  src_dir="/etc/letsencrypt/live/i247.com"
  tgt_dir="/apps/math/keys"

  if sudo test -f ${src_dir:-notset}/fullchain.pem
  then
    # echo "syncing t${n} fullchain.pem..."
    echo "syncing ${src_dir}/fullchain.pem..."

    sudo rsync -rLti --chown=mot:mot --chmod=400 ${src_dir}/fullchain.pem ${tgt_dir}/

  fi

  if sudo test -f ${src_dir:-notset}/privkey.pem
  then
    # echo "syncing t${n} privkey.pem..."
    echo "syncing ${src_dir}/privkey.pem..."

    sudo rsync -rLti --chown=mot:mot --chmod=400 ${src_dir}/privkey.pem ${tgt_dir}/

  fi

done
