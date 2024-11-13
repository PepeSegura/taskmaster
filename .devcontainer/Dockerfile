# Use an official Ubuntu as a parent image
FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y zsh build-essential curl git netcat python3 python3-pip && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    pip install -f /workspaces/taskmaster/srcs/requirements.txt

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# Set Zsh as the default shell
RUN chsh -s $(which zsh)

CMD ["zsh"]
