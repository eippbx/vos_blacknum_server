#!/bin/bash
ydate=$(date -d "yesterday" +%Y%m%d)
/usr/bin/nohup /root/vos_search_up -date ${ydate} 1>/var/log/vos_search.log &
