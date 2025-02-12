# Use an official Ubuntu as a parent image
FROM golang:1.23-bookworm

RUN apt-get update && \
    apt-get install -y zsh build-essential curl git python3 python3-pip supervisor btop nginx && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

RUN ln -sf /dev/stdout /var/log/nginx/access.log
RUN ln -sf /dev/stderr /var/log/nginx/error.log

# Set Zsh as the default shell
RUN chsh -s $(which zsh)

CMD ["zsh"]
