# STEP 1 build executable binary
FROM ubuntu 

# Setup vim and golang to make development easier
RUN apt-get update && apt-get install -y \
    golang \
    curl \
    git \
    vim 
RUN  git clone https://github.com/VundleVim/Vundle.vim.git ~/.vim/bundle/Vundle.vim
COPY containerception /usr/local/bin 
COPY vimrc /root/.vimrc

