from rest_framework import serializers
from main.models import User, Board
from main.validation.val_auth import authentication_error, authorization_error
from main.validation.val_custom import CustomAPIException
from main.util import create_board
import bcrypt
import status


class ClientStateSerializer(serializers.Serializer):
    auth_user = serializers.CharField()
    auth_token = serializers.CharField()
    board_id = serializers.IntegerField(default=-1)

    def create(self, validated_data):
        pass

    def update(self, instance, validated_data):
        pass

    def validate(self, attrs):
        try:
            user = User.objects.prefetch_related(
                'team',
                'board_set',
                'team__user_set',
                'team__board_set__user',
                'team__board_set__column_set',
                'team__board_set__column_set__task_set',
                'team__board_set__column_set__task_set__subtask_set'
            ).get(username=attrs.get('auth_user'))
        except User.DoesNotExist:
            raise CustomAPIException('username',
                                     'User not found.',
                                     status.HTTP_400_BAD_REQUEST)

        valid_token = bytes(user.username, 'utf-8') + user.password
        match = bcrypt.checkpw(
            valid_token,
            bytes(attrs.get('auth_token'), 'utf-8')
        )
        if not match:
            raise authentication_error

        if not user.board_set.all() and user.is_admin:
            board, error_response = create_board(name='New Board',
                                                 team_id=user.team_id,
                                                 team_admin=user)
            if error_response:
                return error_response

        board_id = attrs.get('board_id')
        if board_id:
            try:
                board = user.board_set.get(id=board_id)
            except Board.DoesNotExist:
                raise authorization_error
        else:
            board = user.board_set.all().first()

        if not board:
            err_detail = 'Please ask your team admin to add you to a board ' \
                         'and try again.',
            raise CustomAPIException('board',
                                     err_detail,
                                     status.HTTP_400_BAD_REQUEST)

        if board.team_id != user.team_id:
            raise authorization_error

        return {'user': user, 'board': board}

    def to_representation(self, instance):
        user = instance.get('user')
        board = instance.get('board')

        team_members = user.team.user_set.all()
        board_members = board.user.all()
        boards = user.board_set.all()

        return {
            'user': {
                'username': user.username,
                'teamId': user.team_id,
                'isAdmin': user.is_admin,
                'isAuthenticated': True
            },
            'team': user.is_admin and {
                'id': user.team.id,
                'inviteCode': user.team.invite_code
            },
            'boards': [{
                'id': board.id, 'name': board.name
            } for board in boards],
            'activeBoard': {
                'id': board.id,
                'columns': [{
                    'id': column.id,
                    'order': column.order,
                    'tasks': column.task_set is not None and [{
                        'id': task.id,
                        'title': task.title,
                        'description': task.description,
                        'order': task.order,
                        'user': task.user.username if task.user else '',
                        'subtasks': task.subtask_set is not None and [{
                            'id': subtask.id,
                            'title': subtask.title,
                            'order': subtask.order,
                            'done': subtask.done
                        } for subtask in task.subtask_set.all()]
                    } for task in column.task_set.all()]
                } for column in board.column_set.all()]
            },
            'members': [
                {
                    'username': member.username,
                    'isActive': member in board_members,
                    'isAdmin': member.is_admin
                } for member in sorted(
                    team_members,
                    key=lambda member: not member.is_admin
                )
            ]
        }


