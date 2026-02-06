FROM node:20-alpine AS build
ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}
WORKDIR /app
COPY frontend/package.json ./
COPY frontend/package-lock.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

FROM nginx:1.25-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
