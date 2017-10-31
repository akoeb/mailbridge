FROM scratch
ADD mailbridge /
ADD COMMIT /
CMD ["/mailbridge", "-configFile", "/config.json"]