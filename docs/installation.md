# Installation

## Prerequisites

- [go 1.23 or above](https://go.dev/doc/install)
- [slsa-framework/slsa-verifier](https://github.com/slsa-framework/slsa-verifier#installation)
- [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md)
- [cosign](https://docs.sigstore.dev/cosign/system_config/installation/)

## From release

1. Get the latest release version.

```bash
VERSION=$(curl -s "https://api.github.com/repos/minhthong582000/rate-limiter/releases/latest" | jq -r '.tag_name')
```

or set a specific version:

```bash
VERSION=vX.Y.Z   # Version number with a leading v
```

2. Download the binary that matches your OS and architecture.

```bash
## MacOS arm64
OS=darwin
ARCH=arm64
curl -sSfL https://github.com/minhthong582000/rate-limiter/releases/download/$VERSION/rate-limiter-$OS-$ARCH -o rate-limiter

## Linux amd64
OS=darwin
ARCH=arm64
curl -sSfL https://github.com/minhthong582000/rate-limiter/releases/download/$VERSION/rate-limiter-$OS-$ARCH -o rate-limiter
```

3. Verify the signature. This repository's CI generate [SLSA 3 provenance](https://slsa.dev) using the OpenSSF's [slsa-framework/slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator). To verify our release, install the verification tool from [slsa-framework/slsa-verifier#installation](https://github.com/slsa-framework/slsa-verifier#installation) and verify as follows:

```bash
curl -sL https://github.com/minhthong582000/rate-limiter/releases/download/$VERSION/rate-limiter.intoto.jsonl > provenance.intoto.jsonl

# NOTE: You may be using a different architecture.
slsa-verifier verify-artifact rate-limiter \
  --provenance-path provenance.intoto.jsonl \
  --source-uri github.com/minhthong582000/rate-limiter \
  --source-tag "${VERSION}"

# Verifying artifact rate-limiter: PASSED
# PASSED: SLSA verification passed
```

4. Make the binary executable.

```bash
chmod +x rate-limiter
```

5. Run the binary.

```bash
./rate-limiter --help
```

## From Docker

1. Get the latest release version.

```bash
VERSION=$(curl -s "https://api.github.com/repos/minhthong582000/rate-limiter/releases/latest" | jq -r '.tag_name')
```

or set a specific version:

```bash
VERSION=vX.Y.Z   # Version number with a leading v
```

2. Pull the Docker image.

```bash
docker pull ghcr.io/minhthong582000/rate-limiter:${VERSION}
```

3. Verify the image. This repository's CI generate [SLSA 3 provenance](https://slsa.dev) using the OpenSSF's [slsa-framework/slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator). To verify the image, install the verification tool from [slsa-framework/slsa-verifier#installation](https://github.com/slsa-framework/slsa-verifier#installation) and verify as follows:

```bash
IMAGE=ghcr.io/minhthong582000/rate-limiter:${VERSION}
IMAGE="${IMAGE}@"$(crane digest "${IMAGE}")
slsa-verifier verify-image "$IMAGE" \
    --source-uri github.com/minhthong582000/rate-limiter \
    --source-tag "${VERSION}"

# PASSED: SLSA verification passed
```

4. Verify image. Images are signed by [cosign](https://github.com/sigstore/cosign) using identity-based ("keyless") signing and transparency. Executing the following command to verify the signature of a container image:

```bash
cosign verify \
  --certificate-identity-regexp https://github.com/minhthong582000/rate-limiter/.github/workflows/release.yaml@refs/tags/v \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-github-workflow-repository "minhthong582000/rate-limiter" \
  ghcr.io/minhthong582000/rate-limiter:${VERSION} | jq

# Verification for ghcr.io/minhthong582000/rate-limiter:vx.y.z --
# The following checks were performed on each of these signatures:
#   - The cosign claims were validated
#   - Existence of the claims in the transparency log was verified offline
#   - The code-signing certificate was verified using trusted certificate authority certificates
# [
#   {
#     "critical": {
#       "identity": {
#         "docker-reference": "ghcr.io/minhthong582000/rate-limiter"
#       },
#       "image": {
#         "docker-manifest-digest": "sha256:abcxyz..."
#       },
#       "type": "cosign container image signature"
#     },
#     "optional": {
#       "1.3.6.1.4.1.57264.1.1": "https://token.actions.githubusercontent.com",
#       "1.3.6.1.4.1.57264.1.2": "push",
#       "1.3.6.1.4.1.57264.1.3": "abc123...",
#       "1.3.6.1.4.1.57264.1.4": "Release",
#       "1.3.6.1.4.1.57264.1.5": "minhthong582000/rate-limiter",
#       "1.3.6.1.4.1.57264.1.6": "refs/tags/vx.y.z",
#       "Bundle": {
```

5. Run the Docker image.

```bash
docker run --rm -it ghcr.io/minhthong582000/rate-limiter:${VERSION} --help
```

## From source

```bash
make build
./bin/rate-limiter --help
```
