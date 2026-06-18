---
title: Arquitectura del TUI
description: Arquitectura de la interfaz terminal de Stellar Explorer y su ruta de datos dedicada.
---

## Descripción general

El TUI de Stellar Explorer es una interfaz terminal para inspeccionar datos de la red Stellar. Está diseñado para usuarios que necesitan acceso rápido, con teclado, a ledgers, transacciones, cuentas, activos, contratos y actividad Soroban desde la línea de comandos.

El TUI se compone de dos partes coordinadas:

- `apps/tui`: aplicación terminal escrita en Go, responsable de la experiencia de usuario, estado local, navegación y acceso directo a Stellar RPC.
- `services/tui-indexer`: backend de Stellar Explorer dedicado a flujos terminales, responsable de lecturas indexadas, búsqueda, timelines y datos de actividad en vivo.

Esta estructura mantiene el producto terminal enfocado: la aplicación conserva una experiencia local y rápida, mientras el backend prepara datos más ricos para vistas que necesitan más contexto que una consulta individual a Stellar RPC.

## Estado actual

El TUI es una alpha funcional que cubre los siete tracks de implementación del roadmap técnico con madurez de primera release. Puede correr como cliente terminal directo contra Stellar RPC, y puede usar `services/tui-indexer` para lecturas indexadas más ricas en modo híbrido o indexado.

Las superficies implementadas incluyen:

- navegación conectada con breadcrumbs, persistencia de selección y sublistas de exploración
- effects de transacción, vistas de detalle de operación y recorrido de entidades relacionadas
- detalle de invocaciones, eventos y storage Soroban con modos decoded/raw
- live feed nativo de streams con ingesta Redis, fallback por polling, replay, filtros y presets de watch
- flujos editables de workspace local para bookmarks, labels, notas, vistas guardadas, fallback de cache y aislamiento por perfil
- búsqueda agrupada con filtros de metadata local, ranking y paginación del backend
- cobertura de confiabilidad con tiers de tests unitarios, fixtures, integración y read APIs del indexer

Los roadmaps de producto y técnicos siguen describiendo pulido y profundidad adicional más allá de esta alpha. Deben leerse como dirección planificada, no como garantía de que cada punto futuro ya esté completo.

## Limitaciones conocidas

- El TUI aún no es una release estable; el comportamiento de comandos y la composición de vistas pueden cambiar.
- El modo RPC es intencionalmente más limitado que el modo indexado.
- Los flujos híbridos/indexados requieren `services/tui-indexer` y su infraestructura local.
- Los filtros avanzados del live feed por contrato, activo y tipo de operación dependen de campos opcionales de metadata indexada.
- La ingesta por stream Redis requiere `redis_url` en el perfil activo y un publicador live de `tui-indexer` en ejecución.
- Las vistas guardadas restauran comandos y contexto de pantalla; no replican todo el estado de selección de la UI más allá de los filtros almacenados.
- Algunas secciones indexadas no disponibles se muestran como degradadas en vez de mezclar silenciosamente payloads RPC e indexados.

## Principios de diseño

- **Experiencia terminal primero:** navegación con teclado, vistas compactas, búsqueda por comandos y atajos consistentes.
- **Modelo de datos propio de Stellar:** ledgers, transacciones, operaciones, cuentas, activos, contratos, trustlines, eventos Soroban y metadatos de Stellar RPC son conceptos principales.
- **Contexto local persistente:** perfiles, etiquetas, marcadores, notas y estado de sesión se guardan localmente para continuar investigaciones entre sesiones.
- **Lecturas con origen visible:** la interfaz muestra si un resultado viene de Stellar RPC, del backend del TUI o de una ruta de fallback.
- **Profundidad progresiva:** el TUI funciona con Stellar RPC directo y gana búsqueda, listas, timelines y entidades relacionadas cuando `services/tui-indexer` está disponible.
- **Resolución de una sola fuente:** cada operación de lectura se resuelve desde una fuente, evitando mezclar campos de payloads distintos en un mismo resultado.

## Modelo de ejecución

`apps/tui` usa Bubble Tea para el runtime interactivo y Lip Gloss para renderizar la interfaz. El shell se organiza en regiones estables:

- encabezado con perfil, red y contexto de fuente de datos
- navegación lateral para pantallas principales
- área principal para live feed, búsqueda y vistas de detalle
- área de estado para ciclo de vida y fallback
- overlays de ayuda y búsqueda por comandos

Cada vista conserva su propio estado de interacción. Live feed, lookup, settings, home y command palette administran su selección, scroll y estados de carga, vacío y error. Esto mantiene un comportamiento predecible a medida que se agregan vistas más profundas.

## Modos de datos

### Modo RPC

El modo RPC consulta Stellar RPC directamente. Permite iniciar rápido y usar flujos básicos sin levantar servicios locales adicionales.

Casos típicos:

- inspeccionar actividad reciente
- buscar un ledger, transacción, cuenta o contrato
- validar metadatos de cuentas y contratos desde Stellar RPC

### Modo indexado

El modo indexado lee desde `services/tui-indexer`. Se usa cuando el terminal necesita datos indexados, listas, timelines, entidades relacionadas, holders, operaciones, eventos y búsqueda.

Casos típicos:

- navegar ledgers, cuentas, activos y contratos recientes
- inspeccionar timelines de cuentas, activos y contratos
- buscar entidades Stellar indexadas
- abrir transacciones, operaciones, holders y eventos relacionados

### Modo híbrido

El modo híbrido prefiere `services/tui-indexer` y cae a Stellar RPC cuando el backend no puede servir una operación. El TUI muestra la fuente resuelta y el motivo del fallback para que el usuario entienda el nivel de completitud de la vista.

## Persistencia local

El TUI guarda estado de usuario en SQLite:

- perfiles
- marcadores
- etiquetas
- notas
- estado de sesión
- adjuntos de metadatos locales

Esto convierte al TUI en un espacio de trabajo persistente, no solo en un cliente temporal de consultas. La misma capa local permite futuras vistas guardadas, revisita de entidades inspeccionadas y contexto operativo asociado a objetos de Stellar.

## Responsabilidades del backend

`services/tui-indexer` prepara datos para exploración terminal:

- ingerir ledgers, transacciones, operaciones, activos, cuentas, contratos, eventos de tokens y eventos de contratos
- exponer APIs de lectura para lookup, listas, sublistas y timelines
- proveer búsqueda sobre entidades indexadas de Stellar
- publicar datos de actividad en vivo para monitoreo terminal
- normalizar registros relacionados para que el TUI pueda moverse rápido entre entidades

## Beneficios

- investigación más rápida desde la línea de comandos
- menos ciclos manuales de copiar y consultar
- visibilidad clara de fuente activa y fallback
- modo inicial útil con Stellar RPC directo
- exploración más rica cuando hay datos indexados de Stellar Explorer
- contexto local persistente para análisis recurrentes

## Ruta de madurez

1. Mantener el modo RPC directo útil para exploración básica.
2. Expandir lecturas indexadas para listas, timelines y datos relacionados.
3. Profundizar decode Soroban para invocaciones, eventos, specs y storage.
4. Agregar monitoreo en vivo nativo de streams.
5. Ampliar caché local y flujos de investigación guardada.

Para la dirección de producto planificada, consulta el [Roadmap de producto del TUI](./tui-product-roadmap/). Para las líneas de implementación, consulta el [Roadmap técnico del TUI](./tui-technical-roadmap/).
