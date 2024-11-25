FROM registry.corp.kuaishou.com/kuaishou/centos7.3:latest_web_server

COPY ./A/output /home/web_server/A

WORKDIR /home/web_server/A

# 启动命令
ENV START_COMMAND /home/web_server/A/bin/start.sh
# 停止命令，优雅退出需要
ENV STOP_COMMAND /home/web_server/A/bin/stop.sh

# 升级kcsize
RUN rm -rf /usr/local/sbin/kcsize && rm -rf /usr/local/bin/kcsize && curl -o /usr/local/bin/kcsize http://files.kuaishou.internal:8901/data/repo/sources/kuaishou/public/private_clouds/7331/kcsize && chmod a+x /usr/local/bin/kcsize && ln -s /usr/local/bin/kcsize /usr/local/sbin/kcsize

#启动命令
CMD ["/usr/local/bin/kcsize", "api"]
# CMD ["/home/web_server/A/bin/start.sh"]
