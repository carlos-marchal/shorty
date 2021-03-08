FROM alpine/git
RUN adduser -D git
USER git
COPY --chown=git ./repos /home/git
COPY --chown=git ./ssh/server/server /home/git/.ssh/host
RUN chmod 600 /home/git/.ssh/host
COPY --chown=git ./ssh/client/client.pub /home/git/.ssh/authorized_keys
EXPOSE 22
ENTRYPOINT ["/usr/sbin/sshd", "-D", "-h", "/home/git/.ssh/host"]