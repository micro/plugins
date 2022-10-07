#!/bin/bash

GO_TEST_FLAGS="-v -race -cover -bench=."

# Find directories that contain changes
function find_changes() {
	# Pull main branch
	git checkout main &>/dev/null
	git checkout $GITHUB_REF_NAME &>/dev/null

	# Find all directories that have changed files
	hash=$(git merge-base --fork-point main)
	changes=$(git diff --name-only $hash |
		xargs -d'\n' -I{} dirname {} | sort -u)

	changes=$(find $(echo $changes) -name 'go.mod' -printf '%h\n')

	echo $changes
}

# Find all go directories
function find_all() {
	find . -name 'go.mod' -printf '%h\n'
}

# Run GoLangCi Linters
function run_linter() {
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.0

	golangci-lint --version

	failed="false"

	cwd=$(pwd)
	for dir in $(echo $1); do
		echo "Running linter on $dir"
		pushd $dir >/dev/null

		golangci-lint run -c "${cwd}/.golangci.yaml"

		# Keep track of exit code of linter
		if [[ $? -ne 0 ]]; then
			failed="true"
		fi

		popd >/dev/null
	done

	if [[ $failed == "true" ]]; then
		echo "Linter failed"
		exit 1
	fi
}

# Run unit tests with tparse to create a summary
function create_summary() {
	go install github.com/mfridman/tparse@latest

	cwd=$(pwd)
	for dir in $(echo $1); do
		echo "Creating summary for $dir"
		pushd $dir >/dev/null

		# Download all modules
		go get -v -t -d ./...

		go test $GO_TEST_FLAGS -json ./... |
			tparse -notests -format=markdown >>$GITHUB_STEP_SUMMARY
		popd >/dev/null
	done
}

# Run Unit tests with RichGo for pretty output
function run_test() {
	go install github.com/kyoh86/richgo@latest

	cwd=$(pwd)
	for dir in $(echo $1); do
		echo "Creating summary for $dir"
		pushd $dir >/dev/null

		# Download all modules
		go get -v -t -d ./...

		richgo test $GO_TEST_FLAGS ./...
		popd >/dev/null
	done
}

# Get the dir list based on command type
function get_dirs() {
	if [[ $1 == "all" ]]; then
		find_all
	else
		find_changes
	fi
}

echo "Using branch: $GITHUB_REF_NAME"
case $1 in
"lint")
	dirs=$(get_dirs $2)

	echo "Found $(echo $dirs | wc -l) changed directories"
	echo "Changed dirs:"
	echo $dirs

	run_linter $dirs
	;;
"test")
	dirs=$(get_dirs $2)
	run_test $dirs
	;;
"summary")
	dirs=$(get_dirs $2)
	create_summary $dirs
	;;
"")
	echo "Please provider a command"
	exit 1
	;;
*)
	echo "Command not found: $1"
	exit 1
	;;
esac
