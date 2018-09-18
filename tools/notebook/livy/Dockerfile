# VERSION 1
# Author: Tim.Xu
# Command format: Instruction [arguments / command] ..
# Base image to use, this must be set as the first line
FROM spark-hadoop-base

# Maintainer: docker_user <docker_user at email.com> (@docker_user)
LABEL maintainer="xiaoxubeii@gmail.com"

ENV LIVY_HOME /livy
ENV LIBDIR ${LIVY_HOME}/jars
ENV LIVY_CLASSPATH $LIBDIR/*:$LIVY_CONF_DIR
ENV SPARK_CONF_DIR ${SPARK_HOME}/conf
ENV HADOOP_CONF_DIR ${HADOOP_HOME}/etc
ENV  LIVY_CLASSPATH ${LIVY_CLASSPATH}:${SPARK_CONF_DIR}:${HADOOP_CONF_DIR}
ENTRYPOINT [ "/sbin/tini", "--" ]
WORKDIR ${LIVY_HOME}
CMD ["${JAVA_HOME}/bin/java", "-cp", "${LIVY_CLASSPATH}:$CLASSPATH", "org.apache.livy.server.LivyServer"]