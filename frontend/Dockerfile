# 베이스 이미지
FROM node:20-alpine

WORKDIR /app

# 패키지 파일 복사 및 의존성 설치
COPY package.json package-lock.json ./
RUN npm ci

# 소스 코드 및 설정 파일 복사
COPY src ./src
COPY components.json ./components.json
COPY postcss.config.mjs ./postcss.config.mjs
COPY next-env.d.ts ./next-env.d.ts
COPY next.config.ts ./next.config.ts
COPY tsconfig.json ./tsconfig.json
COPY eslint.config.mjs ./eslint.config.mjs

RUN npm run build

CMD ["npm", "start"]
