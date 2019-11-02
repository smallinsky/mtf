FROM google/cloud-sdk:alpine

RUN apk --no-cache add openjdk8-jre
RUN gcloud components install --quiet beta pubsub-emulator
RUN mkdir -p /var/pubsub

VOLUME /var/pubsub

EXPOSE 8085

ENTRYPOINT ["gcloud", "beta", "emulators", "pubsub"]
CMD ["start", "--host-port=0.0.0.0:8085", "--data-dir=/var/pubsub", "--log-http", "--verbosity=debug", "--user-output-enabled"]
