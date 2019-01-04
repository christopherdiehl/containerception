# STEP 1 build executable binary
FROM ubuntu 

# Setup vim and golang to make development easier
RUN apt-get update && apt-get install -y \
    golang \
    curl \
    git \
    vim \
    wget 
RUN  git clone https://github.com/VundleVim/Vundle.vim.git ~/.vim/bundle/Vundle.vim
COPY containerception /usr/local/bin 
COPY vimrc /root/.vimrc
#setup proc and other needed folders
RUN mkdir rootfs
RUN mkdir rootfs/proc
RUN wget "https://raw.githubusercontent.com/teddyking/ns-process/4.0/assets/busybox.tar"
RUN tar -xf busybox.tar rootfs/
RUN cp /root/bin/bash /root/rootfs/bin/bash