# Security Policy - *Política de seguridad*

## Supported Versions

Only the latest version published on [Releases](https://github.com/inusui/obsidian-termsidian/releases) receives security fixes. If you're using an older version, update before reporting.

*Versiones con soporte*

*Solo la última versión publicada en [Releases](https://github.com/inusui/obsidian-termsidian/releases) recibe arreglos de seguridad. Si usas una versión anterior, actualiza antes de reportar.*

## How to Report a Vulnerability

**Don't open a public issue:** anyone could exploit the vulnerability before it's fixed.

Report it privately from the repository's **Security** tab, using the **Report a vulnerability** button, or directly at [this link](https://github.com/inusui/obsidian-termsidian/security/advisories/new). Only you and I will see the report.

If possible, include:

- the Termsidian version and your operating system;
- what happens and what it allows an attacker to do;
- steps to reproduce it.

You can write in English or Spanish.

*Cómo reportar una vulnerabilidad*

*No abras un issue público:* *cualquiera podría aprovechar la vulnerabilidad antes de que esté arreglada.*

*Repórtala en privado desde la pestaña* **Security** *del repositorio, con el botón* **Report a vulnerability***, o directamente en [este enlace](https://github.com/inusui/obsidian-termsidian/security/advisories/new). El reporte solo lo vemos tú y yo.*

*Si puedes, incluye:*

- *la versión de Termsidian y tu sistema operativo;*
- *qué ocurre y qué le permite hacer a un atacante;*
- *los pasos para reproducirlo.*

*Puedes escribir en español o en inglés.*

## What to Expect

Termsidian is a personal project maintained by a single person, so there are no guaranteed timelines. Even so, vulnerabilities come before any other task:

1. I'll confirm that I received the report.
2. If the vulnerability is confirmed, I'll prepare the fix and publish a new version.
3. Afterward I'll publish the security advisory, with your name if you want to be credited.

*Qué esperar*

*Termsidian es un proyecto personal que mantiene una sola persona, así que no hay plazos garantizados. Aun así, las vulnerabilidades van antes que cualquier otra tarea:*

1. *Te confirmaré que recibí el reporte.*
2. *Si la vulnerabilidad se confirma, prepararé el arreglo y publicaré una versión nueva.*
3. *Después publicaré el aviso de seguridad, con tu nombre si quieres aparecer.*

## What Counts as a Vulnerability

Termsidian opens a terminal with your shell and your permissions. Being able to run any command from it is its function, not a vulnerability.

These do count, for example:

- the plugin or the helper executing something the user didn't write;
- a malformed message between the plugin and the helper causing something worse than an error;
- a Release binary not matching the one obtained by compiling the code at its tag.

If the vulnerability is in a dependency (xterm.js or a Go module) and it's already public, you can open a regular issue instead.

*Qué cuenta como vulnerabilidad*

*Termsidian abre una terminal con tu shell y con tus permisos. Poder ejecutar cualquier comando desde ella es su función, no una vulnerabilidad.*

*Sí lo son, por ejemplo:*

- *que el plugin o el helper ejecuten algo que el usuario no escribió;*
- *que un mensaje malformado entre el plugin y el helper provoque algo más que un error;*
- *que un binario de una Release no coincida con el que se obtiene al compilar el código de su etiqueta.*

*Si la vulnerabilidad está en una dependencia (xterm.js o un módulo de Go) y ya es pública, puedes abrir un issue normal.*

## Verifying a Download

Each Release's binaries are built by GitHub Actions, never by hand on a personal machine. To check that your download arrived complete and matches what was published, compute its hash and compare it against `checksums.txt`, which ships in the same Release. If you installed from the zip, the binaries are in `bin/`.

```bash
sha256sum termsidian-pty-windows-amd64.exe
```

On macOS, use `shasum -a 256` instead of `sha256sum`.

*Verificar una descarga*

*Los binarios de cada Release los compila GitHub Actions, nunca un equipo personal. Para comprobar que tu descarga llegó completa y es la que se publicó, calcula su hash y compáralo con el de `checksums.txt`, que va en la misma Release. Si instalaste desde el zip, los binarios están en `bin/`.*

```bash
sha256sum termsidian-pty-windows-amd64.exe
```

*En macOS, usa `shasum -a 256` en lugar de `sha256sum`.*
