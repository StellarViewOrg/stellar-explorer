---
title: Configuración del entorno de desarrollo
description: Cómo configurar un entorno de desarrollo local.
---

## Requisitos previos

- [Bun](https://bun.sh/) (gestor de paquetes y runtime)
- [Docker](https://www.docker.com/) (para servicios de backend, opcional)
- [D2](https://d2lang.com/) (para renderizar diagramas, opcional)

## Inicio rápido

1. **Clonar el repositorio**

```bash
git clone https://github.com/salazarsebas/stellar-explorer.git
cd stellar-explorer
```

2. **Instalar dependencias**

```bash
bun install
```

3. **Iniciar el servidor de desarrollo**

```bash
bun run dev
```

El explorador estará disponible en `http://localhost:3000`.

## Producto terminal

Stellar Explorer incluye una interfaz terminal escrita en Go para búsqueda y monitoreo de Stellar con teclado:

```bash
bun run tui:build
bun run tui:run
bun run tui:test
```

Para modo híbrido con lecturas indexadas:

```bash
bun run tui-indexer:infra:up
bun run tui-indexer:migrate
bun run tui-indexer:run:serve
bun run tui:run:hybrid
```

## Servicios de backend (opcional)

Para la funcionalidad del indexador, inicia los servicios Docker:

```bash
docker compose -f infra/docker-compose.yml up -d
```

Esto inicia PostgreSQL (puerto 54320), Redis (puerto 63790) y Typesense (puerto 18108).

Para la infraestructura dedicada del backend del TUI, usa:

```bash
bun run tui-indexer:infra:up
```

## Variables de entorno

Copia el archivo de ejemplo y ajusta si es necesario:

```bash
cp .env.local.example .env.local
```

## Comandos disponibles

| Comando | Propósito |
|---|---|
| `bun run dev` | Iniciar servidor de desarrollo (puerto 3000) |
| `bun run build` | Build de producción |
| `bun run lint` | Ejecutar ESLint |
| `bun run format` | Formatear código con Prettier |
| `bun run test` | Ejecutar tests |
| `bun run test:watch` | Ejecutar tests en modo observación |
| `bun run tui:build` | Compilar el TUI |
| `bun run tui:test` | Ejecutar suite completa del TUI (unit + integración) |
| `bun run tui:test:unit` | Ejecutar solo tests unitarios y de fixtures del TUI |
| `bun run tui:test:integration` | Ejecutar solo tests de integración del TUI |
| `bun run tui-indexer:build` | Compilar el backend del TUI |
| `bun run tui-indexer:test` | Ejecutar tests locales del indexer del TUI (sin Docker) |
| `bun run tui-indexer:test:all` | Ejecutar suite completa del indexer (requiere Postgres local) |
| `bun run tui-indexer:infra:up` | Iniciar infraestructura local para `tui-indexer` |
