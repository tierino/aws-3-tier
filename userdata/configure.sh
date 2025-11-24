yum update -y
yum install -y docker
systemctl start docker

aws ecr get-login-password --region $REGION \
  | docker login --username AWS --password-stdin $ACCOUNT.dkr.ecr.$REGION.amazonaws.com

docker pull $REPO:latest
docker run -d -p 8080:8080 $REPO:latest