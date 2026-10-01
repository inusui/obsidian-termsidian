# Política de seguridad

## Versiones con soporte

Solo la última versión publicada en [Releases](https://github.com/inusui/obsidian-termsidian/releases) recibe arreglos de seguridad. Si usas una versión anterior, actualiza antes de reportar.

## Cómo reportar una vulnerabilidad

**No abras un issue público:** cualquiera podría aprovechar la vulnerabilidad antes de que esté arreglada.

Repórtala en privado desde la pestaña **Security** del repositorio, con el botón **Report a vulnerability**, o directamente en [este enlace](https://github.com/inusui/obsidian-termsidian/security/advisories/new). El reporte solo lo vemos tú y yo.

Si puedes, incluye:

- la versión de Termsidian y tu sistema operativo;
- qué ocurre y qué le permite hacer a un atacante;
- los pasos para reproducirlo.

Puedes escribir en español o en inglés.

## Qué esperar

Termsidian es un proyecto personal que mantiene una sola persona, así que no hay plazos garantizados. Aun así, las vulnerabilidades van antes que cualquier otra tarea:

1. Te confirmaré que recibí el reporte.
2. Si la vulnerabilidad se confirma, prepararé el arreglo y publicaré una versión nueva.
3. Después publicaré el aviso de seguridad, con tu nombre si quieres aparecer.

## Qué cuenta como vulnerabilidad

Termsidian abre una terminal con tu shell y con tus permisos. Poder ejecutar cualquier comando desde ella es su función, no una vulnerabilidad.

Sí lo son, por ejemplo:

- que el plugin o el helper ejecuten algo que el usuario no escribió;
- que un mensaje malformado entre el plugin y el helper provoque algo más que un error;
- que un binario de una Release no coincida con el que se obtiene al compilar el código de su etiqueta.

Si la vulnerabilidad está en una dependencia (xterm.js o un módulo de Go) y ya es pública, puedes abrir un issue normal.

## Verificar una descarga

Los binarios de cada Release los compila GitHub Actions, nunca un equipo personal. Para comprobar que tu descarga llegó completa y es la que se publicó, calcula su hash y compáralo con el de `checksums.txt`, que va en la misma Release. Si instalaste desde el zip, los binarios están en `bin/`.

```bash
sha256sum termsidian-pty-windows-amd64.exe
```

En macOS, usa `shasum -a 256` en lugar de `sha256sum`.
