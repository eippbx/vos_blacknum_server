#!/bin/bash
ydate=$(date -d "yesterday" +%Y%m%d)
/root/vos_search_up -date ${ydate}
