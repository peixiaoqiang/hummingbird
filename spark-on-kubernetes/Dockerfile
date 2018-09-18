# VERSION 1
# Author: Tim.Xu
# Command format: Instruction [arguments / command] ..
# Base image to use, this must be set as the first line
FROM spark-hadoop-base

# Maintainer: docker_user <docker_user at email.com> (@docker_user)
LABEL Maintainer="xiaoxubeii@gmail.com"

WORKDIR /usr/local/spark/work-dir

ADD ./entrypoint.sh  /opt/
RUN chmod +x /opt/entrypoint.sh
ENTRYPOINT ["/opt/entrypoint.sh"]