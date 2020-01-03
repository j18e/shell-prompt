RELEASE="v0.2.0"
BASE_URL="https://github.com/j18e/shell-prompt/releases/download/$RELEASE"
ARCH="amd64"
OS="darwin"

if [[ "$(uname -s)" != "Darwin" ]]; then
	OS="linux"
fi

if [[ "$(uname -m)" != "x86_64" ]]; then
	ARCH="386"
fi

ARCHIVE="shell-prompt_${RELEASE}_${OS}_${ARCH}.tar.gz"
curl -Lo $ARCHIVE $BASE_URL/$ARCHIVE
tar -xzf $ARCHIVE
rm -f $ARCHIVE

BINARY="shell-prompt"
sudo mv $BINARY /usr/local/bin/$BINARY
