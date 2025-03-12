FROM ubuntu:24.04 AS build

RUN apt update && apt install ldc gcc dub sqlite3 libsqlite3-dev git -y
RUN git clone https://github.com/soulfind-dev/soulfind.git
WORKDIR /soulfind
RUN dub build --debug=db --debug=msg --debug=user

FROM ubuntu:24.04 AS runtime

RUN apt update && apt install sqlite3 ldc -y
COPY --from=build /soulfind/bin/soulfind /bin
EXPOSE 2242/tcp

CMD ["/bin/soulfind"]