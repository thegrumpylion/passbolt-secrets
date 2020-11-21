FROM scratch

COPY pbscontroller /

ENTRYPOINT [ "/pbscontroller" ]