# TP0: Docker + Comunicaciones + Concurrencia

Este repositorio implementa las resoluciones propuestas a los ejercicios del TP0 de la materia Sistemas Distribuidos 1 de la Facultad de Ingenieria, Universidad de Buenos Aires.  
  
A continuación se detallan las soluciones propuestas para cada ejercicio.

# Ejercicio 1
Para este ejercicio se pidio implementar un script que reciba por parametro la cantidad de clientes a generar, y el archivo `.yaml` a modificar para agregar los clientes.  
  
El script se puede correr con el siguiente comando:
```
./generar-compose.sh <archivo.yaml> <cantidad>
```
   
El algoritmo para generar el archivo esta implementado en Python y fue hecho mediante la modificacion y concatenacion de cadenas correspondientes al formato de un archivo `.yaml`.

# Ejercicio 2
En este ejercicio se tuvo que asociar los archivos de configuracion tanto para el cliente (`config.yaml`) como para el servidor (`config.ini`) para que estos no requieran levantar un nuevo container cada vez que se modificasen. Esto fue hecho con la creacion de un volumen y haciendo *bind mount* de los archivos de configuracion en el container.  
  
# Ejercicio 3
Para este ejercicio se tenia que implementar una forma de probar el echo server con la utilizacion de el programa `nc` (netcat). Las condiciones eran no exponer puertos, y no tener que correr `nc` de manera local, lo que significo tener que levantar un contenedor usando una imagen que fuese compatible con el comando `nc` y que este conectado a la misma red que esta conectado el server.
En mi caso en una primera instancia levante un container con la imagen `subfuzin/netcat` pero debido a un par de problemas con los tests y ejecucion misma del container, opte por una imagen de `alpine`, en el cual aplico el siguiente comando en una terminal:
```bash
echo $expected_response | nc -w 10 $server $server_port
```
Donde `expected_response` es la respuesta esperada del servidor, `server` es la direccion del servidor, y `server_port` es el puerto del servidor. Luego de correr este comando se compara la respuesta del servidor con la respuesta esperada, y en base a eso se determina si el test fue exitoso o no.