FROM scratch
ADD shipyard-deploy /shipyard-deploy
ENTRYPOINT ["/shipyard-deploy"]
CMD ["-h"]
