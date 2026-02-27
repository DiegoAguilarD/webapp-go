// Pipeline de despliegue automatizado para plataforma SaaS multi-tenant
// Cada ejecución despliega un nuevo contenedor de cliente con su propia
// base de datos aislada dentro del MariaDB compartido.

pipeline {
    agent any

    parameters {
        string(name: 'CLIENT_NAME', description: 'Nombre del cliente (ej: Acme Corp)')
        string(name: 'SLUG', description: 'Identificador único del cliente, solo letras minúsculas, números y guiones bajos (ej: acme_corp)')
        string(name: 'ADMIN_EMAIL', description: 'Correo electrónico del administrador del cliente')
        password(name: 'ADMIN_PASSWORD', description: 'Contraseña del administrador del cliente')
    }

    environment {
        ANSIBLE_PLAYBOOK = 'ansible/app_deploy.yml'
        ANSIBLE_INVENTORY = 'ansible/inventory.ini'
    }

    stages {
        stage('Validar parámetros') {
            steps {
                script {
                    echo "=== VALIDANDO PARÁMETROS ==="
                    echo "CLIENT_NAME: '${params.CLIENT_NAME}'"
                    echo "SLUG: '${params.SLUG}'"
                    echo "ADMIN_EMAIL: '${params.ADMIN_EMAIL}'"
                    echo "ADMIN_PASSWORD: ${params.ADMIN_PASSWORD ? '[PRESENTE]' : '[VACÍO]'}"
                    
                    if (!params.CLIENT_NAME?.trim()) {
                        error('CLIENT_NAME es obligatorio.')
                    }
                    if (!params.SLUG?.trim()) {
                        error('SLUG es obligatorio.')
                    }
                    if (!(params.SLUG ==~ /^[a-z][a-z0-9_]*$/)) {
                        error('SLUG debe contener solo letras minúsculas, números y guiones bajos, y comenzar con una letra.')
                    }
                    if (!params.ADMIN_EMAIL?.trim()) {
                        error('ADMIN_EMAIL es obligatorio.')
                    }
                    // Para password parameters, no usar .trim() porque es un objeto Secret
                    if (!params.ADMIN_PASSWORD) {
                        error('ADMIN_PASSWORD es obligatorio.')
                    }
                    
                    echo "=== VALIDACIÓN EXITOSA ==="
                }
            }
        }

        stage('Buscar puerto libre') {
            steps {
                script {
                    // Busca el próximo puerto libre a partir de 9000 verificando
                    // los contenedores Docker activos.
                    env.CLIENT_PORT = sh(
                        returnStdout: true,
                        script: '''#!/bin/sh
                            PORT=9000
                            # Obtener puertos en uso por contenedores (formato: 0.0.0.0:XXXX->YYYY/tcp)
                            USED_PORTS=$(docker ps --format '{{.Ports}}' | grep -o '0.0.0.0:[0-9]*' | cut -d: -f2 | sort -n | uniq)
                            
                            while true; do
                                # Verificar si el puerto está en uso
                                if echo "$USED_PORTS" | grep -q "^${PORT}$"; then
                                    PORT=$((PORT + 1))
                                else
                                    echo "$PORT"
                                    break
                                fi
                                
                                # Evitar bucle infinito
                                if [ $PORT -gt 9100 ]; then
                                    echo "9000"
                                    break
                                fi
                            done
                        '''
                    ).trim()
                    echo "Puerto asignado para el cliente '${params.SLUG}': ${env.CLIENT_PORT}"
                }
            }
        }

        stage('Desplegar cliente') {
            steps {
                echo "Desplegando cliente '${params.CLIENT_NAME}' (${params.SLUG}) en puerto ${env.CLIENT_PORT}..."
                ansiblePlaybook(
                    playbook: "${env.ANSIBLE_PLAYBOOK}",
                    inventory: "${env.ANSIBLE_INVENTORY}",
                    extras: "-e client_name='${params.CLIENT_NAME}' " +
                            "-e client_slug='${params.SLUG}' " +
                            "-e admin_email='${params.ADMIN_EMAIL}' " +
                            "-e admin_password='${params.ADMIN_PASSWORD}' " +
                            "-e client_port='${env.CLIENT_PORT}'"
                )
            }
        }
    }

    post {
        success {
            echo """
            =========================================
            Despliegue exitoso
            Cliente:   ${params.CLIENT_NAME}
            Slug:      ${params.SLUG}
            URL:       http://127.0.0.1:${env.CLIENT_PORT}
            =========================================
            """
        }
        failure {
            echo "El despliegue del cliente '${params.SLUG}' ha fallado. Revise los logs."
        }
    }
}
