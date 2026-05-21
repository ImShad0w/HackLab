FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    openssh-server \
    sudo \
    vim \
    python3 \
    gcc \
    cron \
    && rm -rf /var/lib/apt/lists/*

# Setup SSH
RUN mkdir /var/run/sshd
RUN echo 'root:toor' | chpasswd
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Create low-privilege user
RUN useradd -m -s /bin/bash hacker
RUN echo 'hacker:hacker' | chpasswd

# SUID binary challenge
RUN cp /usr/bin/vim /usr/local/bin/vulnerable-suid
RUN chmod 4755 /usr/local/bin/vulnerable-suid

# Writable cron job challenge
RUN echo '* * * * * root /opt/challenges/cron-task.sh' > /etc/crontab
RUN chmod 644 /etc/crontab

# Capabilities challenge
RUN setcap cap_setuid+ep /usr/bin/python3

# PATH hijacking setup
RUN echo '#!/bin/bash\n/usr/bin/id' > /usr/local/bin/check-status
RUN chmod 755 /usr/local/bin/check-status
RUN echo 'PATH=/tmp:$PATH' >> /home/hacker/.bashrc

EXPOSE 22

CMD ["/usr/sbin/sshd", "-D"]
