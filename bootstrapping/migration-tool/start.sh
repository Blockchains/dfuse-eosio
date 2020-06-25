#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mode=
active_pid=

finish() {
    kill -s TERM $active_pid &> /dev/null || true
}

main() {
  current_dir="`pwd`"
  trap "cd \"$current_dir\"" EXIT
  pushd "$ROOT" &> /dev/null

  while getopts "hm:" opt; do
    case $opt in
      h) usage && exit 0;;
      m) mode="$OPTARG";;
      \?) usage_error "Invalid option: -$OPTARG";;
    esac
  done
  shift $((OPTIND-1))

  if [[ $mode == "export" || $mode == "import" ]]; then
    compile_dfuseeos
  fi

  if [[ $mode == "export" ]]; then
    echo "Ensure you have a FluxDB sever app running (use 'fluxdb-reproc-dev1/start.sh')"
    sleep 2

    rm -rf migration-data
    dfuseeos -c migrator.yaml migrate -i 1000 "$@"
  elif [[ $mode == "import" ]]; then
    rm -rf dfuse-data
    WARN="(.*)" DEBUG="(.*booter.*|.*eosio-boot.*)" dfuseeos -c booter.yaml start "$@"
  else
    usage_error "You must specify either '-m export' or '-m import'"
  fi
}

usage_error() {
  message="$1"
  exit_code="$2"

  echo "ERROR: $message"
  echo ""
  usage
  exit ${exit_code:-1}
}

usage() {
  echo "usage: start.sh -m <mode>"
  echo ""
  echo "Start $(basename $ROOT) environment."
  echo ""
  echo "Options"
  echo "   -m export         Peform the export phase of the migration tool"
  echo "   -m import         Perform the import phase of the migration tool"
}

compile_dfuseeos() {
  pushd "$ROOT/../.." &> /dev/null
    go install ./cmd/dfuseeos
    if [[ $? != 0 ]]; then
      exit 1
    fi
  popd &> /dev/null
}

main $@