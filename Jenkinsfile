@Library('github.com/invoca/jenkins-pipeline@master')
def docker = new io.invoca.Docker()

pipeline {
    agent { label 'docker' }
    stages {
        stage('Setup') {
            steps {
                script {
                    imageArgs = [
                        dockerfile: '.',
                        image_name: 'invocaops/vault-to-env',
                    ]
                }
            }
        }
        stage('Build') {
            steps { script { docker.imageBuild(imageArgs) } }
        }
        stage('Push') {
            environment {
                DOCKERHUB_USER = credentials('dockerhub_user')
                DOCKERHUB_PASSWORD = credentials('dockerhub_password')
            }
            steps { script { docker.imageTagPush(imageArgs.image_name) } }
        }
    }

    post { always { notifySlack(currentBuild.result) } }
}
