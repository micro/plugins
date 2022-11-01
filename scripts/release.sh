#!/bin/bash

######################################################################################
# Release a plugin                                                                   #
#                                                                                    #
# Usage:                                                                             #
#   $ release.sh all                                                                 #
#   $ release.sh v4/broker/http                                                      #
#   $ release.sh v4/broker/http,v4/broker/redis                                      #
#   $ release.sh v4/*                                                                #
#                                                                                    #
######################################################################################

CHANGELOG_TEMPLATE="scripts/template/changelog.md"
CHANGELOG_FILE="/tmp/changelog.md"

function increment_minor_version() {
	declare -a part=(${1//\./ })
	part[2]=0
	part[1]=$((part[1] + 1))
	new="${part[*]}"
	echo -e "${new// /.}"
}

function increment_patch_version() {
	declare -a part=(${1//\./ })
	part[2]=$((part[2] + 1))
	new="${part[*]}"
	echo -e "${new// /.}"
}

function remove_prefix() {
	echo "${1//\.\//}"
}

function check_if_changed() {
	local pkg="$1"
	local last_tag=$(git tag --list --sort='-creatordate' "${pkg}/*" | head -n1)
	if [[ "${last_tag}" == "" ]]; then
		echo -e "# No previous tag\n# Run:\ngh release create "${pkg}/v1.0.0" -n 'Initial release'"
		return 1
	fi

	local changes="$(git --no-pager log "${last_tag}..HEAD" --format="%s" "${pkg}")"
	if [[ "${#changes}" == "0" ]]; then
		# echo "# No changes detected in package '${pkg}'"
		return 1
	fi
	return 0
}

function release() {
	if [[ ! -f "${1}/go.mod" ]]; then
		echo "Unknown package '${1}' given."
		return 1
	fi

	local pkg="${1}"
	if ! check_if_changed "${pkg}"; then
		return 1
	fi

	cat $CHANGELOG_TEMPLATE >$CHANGELOG_FILE

	local last_tag=$(git tag --list --sort='-creatordate' "${pkg}/*" | head -n1)

	# Create changelog file
	git log "${last_tag}..HEAD" --format="%s" "${pkg}" |
		xargs -d'\n' -I{} bash -c "echo \"  * {}\" >> $CHANGELOG_FILE"

	declare -a last_tag_split=(${last_tag//\// })

	local v_version=${last_tag_split[-1]}
	local version=${v_version:1}
	# Remove the version from last_tag_split
	unset last_tag_split[-1]

	# Increment minor version if "feat:" commit found, otherwise patch version
	git --no-pager log "${last_tag}..HEAD" --format="%s" "${pkg}/*" | grep -q -E "^feat:"
	if [[ "$?" == "0" ]]; then
		local tmp_new_tag="$(printf "/%s" "${last_tag_split[@]}")/v$(increment_minor_version ${version})"
		local new_tag=${tmp_new_tag:1}
	else
		local tmp_new_tag="$(printf "/%s" "${last_tag_split[@]}")/v$(increment_patch_version ${version})"
		local new_tag=${tmp_new_tag:1}
	fi

	#  echo -e "# Run:\n"
	echo "# Upgrading pkg ${last_tag} >> ${new_tag}"
	# echo "gh release create \"${new_tag}\" --generate-notes --notes-start-tag ${last_tag}"
	gh release create "${new_tag}" --notes-file "${CHANGELOG_FILE}"
}

function release_all() {
	while read -r pkg; do
		pkg=$(remove_prefix "${pkg}")
		if echo "${pkg}" | grep -q "^v2"; then
			continue
		fi
		release "${pkg}"
	done < <(find . -name 'go.mod' -printf "%h\n")
}

function release_specific() {
	set +o noglob
	while read -r pkg; do
		# If path contains a star find all relevant packages
		if echo "${pkg}" | grep -q "\*"; then
			while read -r p; do
				release "$(remove_prefix "${p}")"
			done < <(find $pkg -name 'go.mod' -printf "%h\n")
		else
			release "${pkg}"
		fi
	done < <(echo "${1}" | tr "," "\n")
	# set -o noglob
	# set +o noglob
}

case $1 in
"all")
	release_all
	;;
*)
	release_specific "${1}"
	;;
esac
