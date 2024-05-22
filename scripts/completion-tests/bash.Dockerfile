# Use a minimal base image
FROM alpine:latest

# Install bash, kubectl, and helm
RUN apk --no-cache add bash bash-completion ca-certificates curl && \
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl && \
    curl -LO "https://get.helm.sh/helm-v3.5.3-linux-amd64.tar.gz" && \
    tar -xzf helm-v3.5.3-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/helm && \
    rm -rf linux-amd64 helm-v3.5.3-linux-amd64.tar.gz

# Copy the Go binary
COPY bbctl_linux /usr/local/bin/bbctl

# Set bash as default shell
ENV SHELL /bin/bash

# Run command to generate configuration file
RUN mkdir -p /root/.config/bbctl && \
    echo "big-bang-repo: \$(bbctl config big-bang-repo)" > /root/.config/bbctl/config.yaml

# Run bash by default when the container starts
CMD ["bash"]

