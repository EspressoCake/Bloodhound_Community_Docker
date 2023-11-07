#!/usr/bin/env python3.11
import docker
import re
import sys

def main():
    if len(sys.argv) != 2:
        raise ValueError('An argument must be supplied')
    else:
        needle = sys.argv[1]

    client = docker.from_env()
    running_instances = [container for container in client.containers.list() if re.findall(f'neo4j-inst-{needle}.*', container.name)]
    if running_instances:
        remotePortForwardString = []
        for instance in running_instances:
            if re.findall('bloodhound', instance.name):
                print(f'Bloodhound browser UI user:  admin')
                print(f'Bloodhound browser password: {re.findall("(?>Initial Password Set To:    )(.*)(?>    #)", instance.logs(stdout=False, stderr=True).decode("UTF-8"))[0]}')
                print(f'Bloodhound browser UI port:  {instance.attrs["NetworkSettings"]["Ports"]["8080/tcp"][0]["HostPort"]}')

                remotePortForwardString.append(instance.attrs["NetworkSettings"]["Ports"]["8080/tcp"][0]["HostPort"])

            elif re.findall('graph-db', instance.name):
                print(f'Neo4j database port:         {instance.attrs["NetworkSettings"]["Ports"]["7687/tcp"][0]["HostPort"]}')
                remotePortForwardString.append(instance.attrs["NetworkSettings"]["Ports"]["7687/tcp"][0]["HostPort"])

                print(f'Neo4j browser port:          {instance.attrs["NetworkSettings"]["Ports"]["7474/tcp"][0]["HostPort"]}')
                remotePortForwardString.append(instance.attrs["NetworkSettings"]["Ports"]["7474/tcp"][0]["HostPort"])

            else:
                pass

        if remotePortForwardString:
            remotePortForwardString = ' -L '.join([f'{port}:localhost:{port}' for port in remotePortForwardString])
            print(f"SSH Port-forwarding:         ssh -L {remotePortForwardString} your_user@remote_system_housing_docker_instances")

if __name__ == '__main__':
    main()
