#!/bin/bash

set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root"

for participants in 2 8 16 30 50; do
  echo "SFU routing profile: ${participants} participants"
  SFU_LOAD_PARTICIPANTS="$participants" go test ./apps/streaming/sfu \
    -run Test_LoadProfile_BoundsFanout \
    -bench BenchmarkRoomFanout \
    -benchmem \
    -benchtime 2s \
    -count 1
done
