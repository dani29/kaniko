FROM busybox

RUN addgroup -g 1001 -S appgroup \
  && adduser -u 1005 -S al -G appgroup \
  && adduser -u 1004 -S bob -G appgroup

# Add local file with and without chown
ADD --chown=1005:appgroup a.txt /a.txt
ADD b.txt /b.txt

# Add remote file with and without chown
ADD --chown=bob:1001 https://raw.githubusercontent.com/GoogleContainerTools/kaniko/master/README.md /r1.txt
ADD https://raw.githubusercontent.com/GoogleContainerTools/kaniko/master/README.md /r2.txt
