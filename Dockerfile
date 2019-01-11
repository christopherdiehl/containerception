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
RUN cp /bin/ -R rootfs/bin
RUN cp /lib64/ -R rootfs/lib64
RUN cp /lib/ -R rootfs/lib
RUN cp /dev/ -R rootfs/dev
RUN cp /mnt/ -R rootfs/mnt
RUN cp /root/ -R rootfs/root
RUN cp /tmp/ -R rootfs/tmp
RUN mkdir rootfs/usr
RUN mkdir rootfs/usr/sbin
