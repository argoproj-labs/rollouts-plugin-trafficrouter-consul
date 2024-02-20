function status {
	if test "${COLORIZE}" -eq 1; then
		tput bold
		tput setaf 4
	fi

	echo "$@"

	if test "${COLORIZE}" -eq 1; then
		tput sgr0
	fi
}

function status_stage {
	if test "${COLORIZE}" -eq 1; then
		tput bold
		tput setaf 2
	fi

	echo "$@"

	if test "${COLORIZE}" -eq 1; then
		tput sgr0
	fi
}

function is_set {
	# Arguments:
	#   $1 - string value to check its truthiness
	#
	# Return:
	#   0 - is truthy (backwards I know but allows syntax like `if is_set <var>` to work)
	#   1 - is not truthy

	local val=$(tr '[:upper:]' '[:lower:]' <<<"$1")
	case $val in
	1 | t | true | y | yes)
		return 0
		;;
	*)
		return 1
		;;
	esac
}

function sed_i {
	if test "$(uname)" == "Darwin"; then
		sed -i '' "$@"
		return $?
	else
		sed -i "$@"
		return $?
	fi
}

function prepare_dev {
	# Arguments:
	#   $1 - Path to top level Consul K8s source
	#   $2 - The version of the release
	#
	# Returns:
	#   0 - success
	#   * - error

  local curDir=$1
  local nextReleaseVersion=$2

	echo "prepare_dev: dir:${curDir} plugin:${nextReleaseVersion} mode:dev"
	set_version "${curDir}" "${nextReleaseVersion}" "dev"

	return 0
}

function prepare_release {
	# Arguments:
	#   $1 - Path to top level Consul K8s source
	#   $2 - The version of the release
	#   $3 - The pre-release version
	#
	#
	# Returns:
	#   0 - success
	#   * - error

  local curDir=$1
  local version=$2
  local prereleaseVersion=$3

	echo "prepare_release: dir:${curDir} plugin:${version} prerelease_version (can be empty):${prereleaseVersion}"
	set_version "${curDir}" "${version}" "${prereleaseVersion}"
}

function set_version {
	# Arguments:
	#   $1 - Path to top level Consul K8s source
	#   $2 - The version of the release
	#   $3 - The pre-release version
	#
	#
	# Returns:
	#   0 - success
	#   * - error

	if ! test -d "$1"; then
		err "ERROR: '$1' is not a directory. prepare_release must be called with the path to a git repo as the first argument"
		return 1
	fi

	if test -z "$2"; then
		err "ERROR: The version specified was empty"
		return 1
	fi

	local sdir="$1"
	local vers="$2"
	local prevers="$3"

	status_stage "==> Updating "${sdir}/pkg/version/version.go" with version info: ${vers} ${prevers}"
	if ! update_version "${sdir}/pkg/version/version.go" "${vers}" "${prevers}"; then
		return 1
	fi

	return 0
}

function update_version {
    # Arguments:
    #   $1 - Path to the version file
    #   $2 - Version string
    #   $3 - PreRelease version (if unset will become an empty string)
    #
    # Returns:
    #   0 - success
    #   * - error

    if ! test -f "$1"; then
        err "ERROR: '$1' is not a regular file. update_version must be called with the path to a go version file"
        return 1
    fi

    if test -z "$2"; then
        err "ERROR: The version specified was empty"
        return 1
    fi

    local vfile="$1"
    local version="$2"
    local prerelease="$3"

    sed_i ${SED_EXT} -e "s/\(Version[[:space:]]*=[[:space:]]*\)\"[^\"]*\"/\1\"${version}\"/g" -e "s/\(VersionPrerelease[[:space:]]*=[[:space:]]*\)\"[^\"]*\"/\1\"${prerelease}\"/g" "${vfile}"
    return $?
}