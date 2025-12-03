#!/usr/bin/env bash

# copy any updated certs from letsencrypt updates

pwd

for n in {0..5}
do

  src_dir="/etc/letsencrypt/live/a1.i247.com"
  tgt_dir="/apps/math/keys"

  if sudo test -f ${src_dir:-notset}/fullchain.pem
  then
    echo "syncing t${n} fullchain.pem..."

    sudo rsync -rLti --chown=mot:mot --chmod=400 ${src_dir}/fullchain.pem ${tgt_dir}/

  fi

  if sudo test -f ${src_dir:-notset}/privkey.pem
  then
    echo "syncing t${n} privkey.pem..."

    sudo rsync -rLti --chown=mot:mot --chmod=400 ${src_dir}/privkey.pem ${tgt_dir}/

  fi

done