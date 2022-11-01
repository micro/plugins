#!/bin/bash

######################################################################################
# $ release.sh broker/http                                                           #
#                                                                                    #
# Release a plugin                                                                   #
#                                                                                    #
######################################################################################

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
	local last_tag=$(git tag --list --sort='-creatordate' "${pkg}/*" | head -n1)
	local changes="$(git --no-pager log "${last_tag}..HEAD" --format="%s" "${pkg}")"

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
	gh release create "${new_tag}" --generate-notes --notes-start-tag "${last_tag}"
}

function release_all() {
	for pkg in $(find . -name 'go.mod' -printf "%h\n"); do
		pkg=$(remove_prefix "${pkg}")
		if echo "${pkg}" | grep -q "^v2"; then
			continue
		fi

		release "${pkg}"
	done
}

function release_specific() {
	set -o noglob
	for pkg in $(echo "${1}" | tr "," "\n"); do
		set +o noglob
		echo "checking: ${pkg}"

		# If path contains a star find all relevant packages
		if echo "${pkg}" | grep -q "\*"; then
			for p in $(find ${pkg} -name 'go.mod' -printf "%h\n"); do
				release "$(remove_prefix "${p}")"
			done
		else
			release "${pkg}"
		fi
	done

}

case $1 in
"all")
	release_all
	;;
*)
	release_specific "${1}"
	;;
esac
