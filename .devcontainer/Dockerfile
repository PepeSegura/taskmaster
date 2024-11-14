# Use an official Ubuntu as a parent image
FROM golang:1.23-bookworm

RUN apt-get update && \
    apt-get install -y zsh build-essential curl git netcat-traditional python3 python3-pip supervisor && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# Set Zsh as the default shell
RUN chsh -s $(which zsh)

CMD ["zsh"]
