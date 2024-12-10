FROM node:18 AS build

WORKDIR /app
COPY ./client/package.json ./client/package-lock.json ./
RUN npm install
COPY ./client ./
RUN npm run build

FROM node:18-slim

WORKDIR /app
COPY --from=build /app ./
EXPOSE 3000
CMD ["npm", "start"]
